package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// PurchaseDocumentRepository persists purchase documents and vendor payments.
type PurchaseDocumentRepository interface {
	ListDocuments(ctx context.Context, kind string, q models.PurchaseDocumentListQuery) ([]models.Document, error)
	GetDocument(ctx context.Context, id int32, kind string, onlyHeader bool) (*models.Document, error)
	CreateDocument(ctx context.Context, kind string, doc *models.Document) (*models.Document, error)
	UpdateDocument(ctx context.Context, id int32, kind string, doc *models.Document) (*models.Document, error)
	DeleteDocument(ctx context.Context, id int32, kind string) error
	TransformDocumentList(ctx context.Context, targetKind string, payload models.DocumentListTransformation) ([]models.Document, error)
	GetPDF(ctx context.Context, id int32) ([]byte, error)
	GetAttachment(ctx context.Context, idDoc int32) (*models.PurchaseDocumentAttachment, error)
	SaveAttachment(ctx context.Context, idDoc int32, filename, contentType string, data []byte) error
	DeleteAttachment(ctx context.Context, idDoc int32) error
	ListVendorPayments(ctx context.Context, q models.ListQuery) ([]models.VendorPayment, error)
	GetVendorPayment(ctx context.Context, id int32) (*models.VendorPayment, error)
	CreateVendorPayment(ctx context.Context, p *models.VendorPayment) (*models.VendorPayment, error)
	UpdateVendorPayment(ctx context.Context, id int32, p *models.VendorPayment) (*models.VendorPayment, error)
	DeleteVendorPayment(ctx context.Context, id int32) error
	GetVendorPaymentByDoc(ctx context.Context, idDoc int32) (*models.VendorPayment, error)
	ListOpenBalance(ctx context.Context, idVendor string, limit, offset int) ([]models.VendorOpenBalanceRow, error)
}

type purchaseDocumentRepository struct {
	db *pgxpool.Pool
}

func NewPurchaseDocumentRepository(db *pgxpool.Pool) PurchaseDocumentRepository {
	return &purchaseDocumentRepository{db: db}
}

const purchaseDocHeaderSQL = `
	SELECT id, COALESCE(id_document_type,0), COALESCE(id_contact,0), COALESCE(id_customer,''), COALESCE(id_vendor,''),
		doc_date, doc_status, COALESCE(doc_currency,''), tdsc1, tdsc2, tdsc3, tdsc4,
		COALESCE(id_bu,''), COALESCE(id_numerator,0), COALESCE(doc_number,''), COALESCE(sequential_number,0),
		COALESCE(note,''), COALESCE(internal_note,''), id_payment_term, COALESCE(reference,''), COALESCE(dest_id_storage,'')
	FROM purchase_documents WHERE company_id = $1 AND deleted = false`

func (r *purchaseDocumentRepository) ListDocuments(ctx context.Context, kind string, q models.PurchaseDocumentListQuery) ([]models.Document, error) {
	sql := purchaseDocHeaderSQL + ` AND doc_kind = $2`
	args := []any{defaultCompanyID, kind}
	n := 3
	if s := strings.TrimSpace(q.FreeText); s != "" {
		sql += fmt.Sprintf(` AND (doc_number ILIKE $%d OR note ILIKE $%d OR id_vendor ILIKE $%d)`, n, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if q.IDContact != nil {
		sql += fmt.Sprintf(` AND id_contact = $%d`, n)
		args = append(args, *q.IDContact)
		n++
	}
	if s := strings.TrimSpace(q.DocStatus); s != "" {
		sql += fmt.Sprintf(` AND doc_status = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.NeDocStatus); s != "" {
		sql += fmt.Sprintf(` AND doc_status <> $%d`, n)
		args = append(args, s)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list purchase documents: %w", err)
	}
	defer rows.Close()
	return scanPurchaseDocumentHeaders(rows)
}

func (r *purchaseDocumentRepository) GetDocument(ctx context.Context, id int32, kind string, onlyHeader bool) (*models.Document, error) {
	sql := purchaseDocHeaderSQL + ` AND id = $2`
	args := []any{defaultCompanyID, id}
	if strings.TrimSpace(kind) != "" {
		sql += ` AND doc_kind = $3`
		args = append(args, kind)
	}
	row := r.db.QueryRow(ctx, sql, args...)
	doc, err := scanPurchaseDocumentHeader(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get purchase document: %w", err)
	}
	if onlyHeader {
		return doc, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT id_pos, COALESCE(id_material,''), COALESCE(id_attribute_combination,0), COALESCE(id_pos_type,1),
			description, price, quantity, COALESCE(packages,1), dsc1, dsc2, dsc3, dsc4, COALESCE(id_vat,''), COALESCE(um,'')
		FROM purchase_document_rows WHERE document_id = $1 ORDER BY id_pos
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row models.DocumentRow
		if err := rows.Scan(&row.IDPos, &row.IDMaterial, &row.IDAttributeCombination, &row.IDPosType,
			&row.Description, &row.Price, &row.Quantity, &row.Packages, &row.Dsc1, &row.Dsc2, &row.Dsc3, &row.Dsc4,
			&row.IDVat, &row.Um); err != nil {
			return nil, err
		}
		doc.Rows = append(doc.Rows, row)
	}
	payments, err := r.loadPayments(ctx, id)
	if err != nil {
		return nil, err
	}
	doc.Payments = payments
	return doc, rows.Err()
}

func (r *purchaseDocumentRepository) CreateDocument(ctx context.Context, kind string, doc *models.Document) (*models.Document, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var seq int64
	if err := tx.QueryRow(ctx, `SELECT nextval('purchase_doc_number_seq')`).Scan(&seq); err != nil {
		return nil, err
	}
	prefix := purchaseDocPrefix(kind)
	docNumber := doc.DocNumber
	if strings.TrimSpace(docNumber) == "" {
		docNumber = fmt.Sprintf("%s-%d", prefix, seq)
	}
	docDate := models.MillisToTime(doc.DocDate)

	var id int32
	err = tx.QueryRow(ctx, `
		INSERT INTO purchase_documents (
			doc_kind, id_document_type, id_contact, id_customer, id_vendor, doc_date, doc_status, doc_currency,
			tdsc1, tdsc2, tdsc3, tdsc4, id_bu, id_numerator, doc_number, sequential_number,
			note, internal_note, id_payment_term, reference, dest_id_storage
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		RETURNING id
	`, kind, nullInt32(doc.IDDocumentType), nullInt32(doc.IDContact), emptyAsNull(doc.IDCustomer), emptyAsNull(doc.IDVendor),
		docDate, defaultStr(doc.DocStatus, "CREATED"), defaultStr(doc.DocCurrency, "EUR"),
		doc.Tdsc1, doc.Tdsc2, doc.Tdsc3, doc.Tdsc4, emptyAsNull(doc.IDBu), nullInt32(doc.IDNumerator),
		docNumber, int32(seq), emptyAsNull(doc.Note), emptyAsNull(doc.InternalNote), doc.IDPaymentTerm,
		emptyAsNull(doc.Reference), emptyAsNull(doc.DestIDStorage)).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert purchase document: %w", err)
	}
	if err := r.insertRowsAndPayments(ctx, tx, id, doc); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetDocument(ctx, id, "", false)
}

func (r *purchaseDocumentRepository) UpdateDocument(ctx context.Context, id int32, kind string, doc *models.Document) (*models.Document, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	sql := `
		UPDATE purchase_documents SET
			id_contact=$2, id_customer=$3, id_vendor=$4, doc_date=$5, doc_status=$6, doc_currency=$7,
			tdsc1=$8, tdsc2=$9, tdsc3=$10, tdsc4=$11, id_bu=$12, id_numerator=$13,
			note=$14, internal_note=$15, id_payment_term=$16, reference=$17, dest_id_storage=$18, updated_at=now()
		WHERE id=$1 AND company_id=$19 AND deleted=false`
	args := []any{
		id, nullInt32(doc.IDContact), emptyAsNull(doc.IDCustomer), emptyAsNull(doc.IDVendor),
		models.MillisToTime(doc.DocDate), defaultStr(doc.DocStatus, "CREATED"), defaultStr(doc.DocCurrency, "EUR"),
		doc.Tdsc1, doc.Tdsc2, doc.Tdsc3, doc.Tdsc4, emptyAsNull(doc.IDBu), nullInt32(doc.IDNumerator),
		emptyAsNull(doc.Note), emptyAsNull(doc.InternalNote), doc.IDPaymentTerm,
		emptyAsNull(doc.Reference), emptyAsNull(doc.DestIDStorage), defaultCompanyID,
	}
	if strings.TrimSpace(kind) != "" {
		sql += ` AND doc_kind = $20`
		args = append(args, kind)
	}
	tag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("purchase document not found")
	}
	if _, err := tx.Exec(ctx, `DELETE FROM purchase_document_rows WHERE document_id=$1`, id); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM purchase_document_payments WHERE document_id=$1`, id); err != nil {
		return nil, err
	}
	if err := r.insertRowsAndPayments(ctx, tx, id, doc); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetDocument(ctx, id, kind, false)
}

func (r *purchaseDocumentRepository) DeleteDocument(ctx context.Context, id int32, kind string) error {
	sql := `UPDATE purchase_documents SET deleted=true, updated_at=now() WHERE id=$1 AND company_id=$2 AND deleted=false`
	args := []any{id, defaultCompanyID}
	if strings.TrimSpace(kind) != "" {
		sql += ` AND doc_kind = $3`
		args = append(args, kind)
	}
	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("purchase document not found")
	}
	return nil
}

func (r *purchaseDocumentRepository) TransformDocument(ctx context.Context, sourceID int32, targetKind string, docDate *int64, docNumber string) (*models.Document, error) {
	src, err := r.GetDocument(ctx, sourceID, "", false)
	if err != nil || src == nil {
		return nil, fmt.Errorf("source document not found")
	}
	if docDate != nil {
		src.DocDate = docDate
	}
	src.DocStatus = "CREATED"
	src.ID = 0
	src.DocNumber = ""
	out, err := r.CreateDocument(ctx, targetKind, src)
	if err != nil {
		return nil, err
	}
	_, _ = r.db.Exec(ctx, `UPDATE purchase_documents SET source_document_id=$2 WHERE id=$1`, out.ID, sourceID)
	if docNumber != "" {
		_, _ = r.db.Exec(ctx, `UPDATE purchase_documents SET doc_number=$2 WHERE id=$1`, out.ID, docNumber)
	}
	return r.GetDocument(ctx, out.ID, "", false)
}

func (r *purchaseDocumentRepository) TransformDocumentList(ctx context.Context, targetKind string, payload models.DocumentListTransformation) ([]models.Document, error) {
	ids := strings.Split(payload.IDDocumentList, ",")
	out := make([]models.Document, 0)
	for _, s := range ids {
		id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 32)
		if err != nil {
			continue
		}
		doc, err := r.TransformDocument(ctx, int32(id), targetKind, payload.DocDate, payload.DocNumber)
		if err != nil {
			return nil, err
		}
		out = append(out, *doc)
	}
	return out, nil
}

func (r *purchaseDocumentRepository) GetPDF(ctx context.Context, id int32) ([]byte, error) {
	var pdf []byte
	err := r.db.QueryRow(ctx, `SELECT pdf_data FROM purchase_documents WHERE id=$1 AND deleted=false`, id).Scan(&pdf)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if len(pdf) > 0 {
		return pdf, nil
	}
	return []byte("%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n"), nil
}

func (r *purchaseDocumentRepository) GetAttachment(ctx context.Context, idDoc int32) (*models.PurchaseDocumentAttachment, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, filename, COALESCE(content_type,'') FROM purchase_document_attachments
		WHERE document_id=$1 ORDER BY id DESC LIMIT 1
	`, idDoc)
	var a models.PurchaseDocumentAttachment
	if err := row.Scan(&a.ID, &a.Filename, &a.ContentType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *purchaseDocumentRepository) SaveAttachment(ctx context.Context, idDoc int32, filename, contentType string, data []byte) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO purchase_document_attachments (document_id, filename, content_type, data) VALUES ($1,$2,$3,$4)
	`, idDoc, filename, contentType, data)
	return err
}

func (r *purchaseDocumentRepository) DeleteAttachment(ctx context.Context, idDoc int32) error {
	_, err := r.db.Exec(ctx, `DELETE FROM purchase_document_attachments WHERE document_id=$1`, idDoc)
	return err
}

func (r *purchaseDocumentRepository) ListVendorPayments(ctx context.Context, q models.ListQuery) ([]models.VendorPayment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, COALESCE(id_document_type,0), COALESCE(reference,''), id_agent,
			total_paid_ic, total_invoice_ic, total_remaining_ic, doc_date, COALESCE(note,''), COALESCE(doc_number,''), deleted
		FROM vendor_payments WHERE company_id=$1 AND deleted=false
		ORDER BY id DESC LIMIT $2 OFFSET $3
	`, defaultCompanyID, limitOrDefault(q.Limit), q.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.VendorPayment, 0)
	for rows.Next() {
		p, err := scanVendorPaymentFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func (r *purchaseDocumentRepository) GetVendorPayment(ctx context.Context, id int32) (*models.VendorPayment, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(id_document_type,0), COALESCE(reference,''), id_agent,
			total_paid_ic, total_invoice_ic, total_remaining_ic, doc_date, COALESCE(note,''), COALESCE(doc_number,''), deleted
		FROM vendor_payments WHERE id=$1 AND company_id=$2 AND deleted=false
	`, id, defaultCompanyID)
	return scanVendorPaymentRow(row)
}

func (r *purchaseDocumentRepository) CreateVendorPayment(ctx context.Context, p *models.VendorPayment) (*models.VendorPayment, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO vendor_payments (id_document_type, reference, id_agent, total_paid_ic, total_invoice_ic, total_remaining_ic, doc_date, note, doc_number)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id
	`, p.IDDocumentType, emptyAsNull(p.Reference), p.IDAgent, p.TotalPaidIc, p.TotalInvoiceIc, p.TotalRemainingIc,
		models.MillisToTime(p.DocDate), emptyAsNull(p.Note), emptyAsNull(p.DocNumber))
	var id int32
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return r.GetVendorPayment(ctx, id)
}

func (r *purchaseDocumentRepository) UpdateVendorPayment(ctx context.Context, id int32, p *models.VendorPayment) (*models.VendorPayment, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE vendor_payments SET reference=$2, total_paid_ic=$3, total_invoice_ic=$4, total_remaining_ic=$5,
			doc_date=$6, note=$7, updated_at=now()
		WHERE id=$1 AND company_id=$8 AND deleted=false
	`, id, emptyAsNull(p.Reference), p.TotalPaidIc, p.TotalInvoiceIc, p.TotalRemainingIc,
		models.MillisToTime(p.DocDate), emptyAsNull(p.Note), defaultCompanyID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("vendor payment not found")
	}
	return r.GetVendorPayment(ctx, id)
}

func (r *purchaseDocumentRepository) DeleteVendorPayment(ctx context.Context, id int32) error {
	tag, err := r.db.Exec(ctx, `UPDATE vendor_payments SET deleted=true WHERE id=$1 AND company_id=$2`, id, defaultCompanyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("vendor payment not found")
	}
	return nil
}

func (r *purchaseDocumentRepository) GetVendorPaymentByDoc(ctx context.Context, idDoc int32) (*models.VendorPayment, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(id_document_type,0), COALESCE(reference,''), id_agent,
			total_paid_ic, total_invoice_ic, total_remaining_ic, doc_date, COALESCE(note,''), COALESCE(doc_number,''), deleted
		FROM vendor_payments WHERE reference=$1 AND company_id=$2 AND deleted=false LIMIT 1
	`, strconv.Itoa(int(idDoc)), defaultCompanyID)
	return scanVendorPaymentRow(row)
}

func (r *purchaseDocumentRepository) ListOpenBalance(ctx context.Context, idVendor string, limit, offset int) ([]models.VendorOpenBalanceRow, error) {
	sql := `
		SELECT COALESCE(d.id_vendor,''), d.id, COALESCE(d.doc_number,''),
			COALESCE(SUM(p.amount),0), COALESCE(SUM(p.paid_amount),0),
			COALESCE(SUM(p.amount),0) - COALESCE(SUM(p.paid_amount),0),
			CASE WHEN COALESCE(SUM(p.amount),0) <= COALESCE(SUM(p.paid_amount),0) THEN 'PAID_STATUS' ELSE 'PENDING_STATUS' END
		FROM purchase_documents d
		LEFT JOIN purchase_document_payments p ON p.document_id = d.id
		WHERE d.doc_kind = 'invoice' AND d.deleted = false AND d.company_id = $1`
	args := []any{defaultCompanyID}
	n := 2
	if s := strings.TrimSpace(idVendor); s != "" {
		sql += fmt.Sprintf(` AND d.id_vendor = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += ` GROUP BY d.id, d.id_vendor, d.doc_number`
	sql += fmt.Sprintf(` ORDER BY d.id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(limit), offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.VendorOpenBalanceRow, 0)
	for rows.Next() {
		var row models.VendorOpenBalanceRow
		if err := rows.Scan(&row.IDVendor, &row.IDDocument, &row.DocNumber, &row.Amount, &row.PaidAmount, &row.Remaining, &row.PaymentStatus); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *purchaseDocumentRepository) insertRowsAndPayments(ctx context.Context, tx pgx.Tx, docID int32, doc *models.Document) error {
	for i, row := range doc.Rows {
		pos := row.IDPos
		if pos == 0 {
			pos = int32(i + 1)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO purchase_document_rows (document_id, id_pos, id_material, id_attribute_combination, id_pos_type,
				description, price, quantity, packages, dsc1, dsc2, dsc3, dsc4, id_vat, um)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		`, docID, pos, emptyAsNull(row.IDMaterial), row.IDAttributeCombination, defaultInt32(row.IDPosType, 1),
			row.Description, row.Price, defaultFloat(row.Quantity, 1), row.Packages,
			row.Dsc1, row.Dsc2, row.Dsc3, row.Dsc4, emptyAsNull(row.IDVat), emptyAsNull(row.Um)); err != nil {
			return err
		}
	}
	for i, p := range doc.Payments {
		rate := p.RateNumber
		if rate == 0 {
			rate = int32(i + 1)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO purchase_document_payments (document_id, rate_number, description, amount, paid_amount, due_date, payment_date)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, docID, rate, emptyAsNull(p.Description), p.Amount, p.PaidAmount,
			millisToTimePtr(p.Date), millisToTimePtr(p.PaymentDate)); err != nil {
			return err
		}
	}
	return nil
}

func (r *purchaseDocumentRepository) loadPayments(ctx context.Context, docID int32) ([]models.DocumentPaymentRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT rate_number, COALESCE(description,''), amount, paid_amount, due_date, payment_date
		FROM purchase_document_payments WHERE document_id=$1 ORDER BY rate_number
	`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.DocumentPaymentRow
	for rows.Next() {
		var p models.DocumentPaymentRow
		var due, paid *time.Time
		if err := rows.Scan(&p.RateNumber, &p.Description, &p.Amount, &p.PaidAmount, &due, &paid); err != nil {
			return nil, err
		}
		if due != nil {
			p.Date = models.MillisPtr(*due)
		}
		if paid != nil {
			p.PaymentDate = models.MillisPtr(*paid)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanPurchaseDocumentHeaders(rows pgx.Rows) ([]models.Document, error) {
	var out []models.Document
	for rows.Next() {
		d, err := scanPurchaseDocumentHeader(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

func scanPurchaseDocumentHeader(s interface{ Scan(...any) error }) (*models.Document, error) {
	var d models.Document
	var docDate time.Time
	var idPaymentTerm *int32
	if err := s.Scan(&d.ID, &d.IDDocumentType, &d.IDContact, &d.IDCustomer, &d.IDVendor,
		&docDate, &d.DocStatus, &d.DocCurrency, &d.Tdsc1, &d.Tdsc2, &d.Tdsc3, &d.Tdsc4,
		&d.IDBu, &d.IDNumerator, &d.DocNumber, &d.SequentialNumber,
		&d.Note, &d.InternalNote, &idPaymentTerm, &d.Reference, &d.DestIDStorage); err != nil {
		return nil, err
	}
	d.DocDate = models.MillisPtr(docDate)
	d.IDPaymentTerm = idPaymentTerm
	return &d, nil
}

func scanVendorPayments(rows pgx.Rows) ([]models.VendorPayment, error) {
	var out []models.VendorPayment
	for rows.Next() {
		p, err := scanVendorPaymentFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func scanVendorPaymentRow(row pgx.Row) (*models.VendorPayment, error) {
	p, err := scanVendorPaymentFromScanner(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func scanVendorPaymentFromScanner(s interface{ Scan(...any) error }) (*models.VendorPayment, error) {
	var p models.VendorPayment
	var docDate time.Time
	var idAgent *int32
	if err := s.Scan(&p.ID, &p.IDDocumentType, &p.Reference, &idAgent,
		&p.TotalPaidIc, &p.TotalInvoiceIc, &p.TotalRemainingIc, &docDate, &p.Note, &p.DocNumber, &p.Deleted); err != nil {
		return nil, err
	}
	p.IDAgent = idAgent
	p.DocDate = models.MillisPtr(docDate)
	return &p, nil
}

func purchaseDocPrefix(kind string) string {
	switch kind {
	case models.PurchaseDocOrder:
		return "PO"
	case models.PurchaseDocGoodsReceipt:
		return "GR"
	case models.PurchaseDocInvoice:
		return "PINV"
	default:
		return strings.ToUpper(kind[:min(3, len(kind))])
	}
}
