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

// SaleDocumentRepository persists sales documents and customer payments.
type SaleDocumentRepository interface {
	ListDocuments(ctx context.Context, kind string, q models.SaleDocumentListQuery) ([]models.Document, error)
	GetDocument(ctx context.Context, id int32, kind string, onlyHeader bool) (*models.Document, error)
	CreateDocument(ctx context.Context, kind string, doc *models.Document) (*models.Document, error)
	UpdateDocument(ctx context.Context, id int32, kind string, doc *models.Document) (*models.Document, error)
	DeleteDocument(ctx context.Context, id int32, kind string) error
	TransformDocument(ctx context.Context, sourceID int32, targetKind string, docDate *int64, docNumber string) (*models.Document, error)
	TransformDocumentList(ctx context.Context, targetKind string, payload models.DocumentListTransformation) ([]models.Document, error)
	GetPDF(ctx context.Context, id int32) ([]byte, error)
	GetAttachment(ctx context.Context, idDoc int32) (*models.SaleDocumentAttachment, error)
	SaveAttachment(ctx context.Context, idDoc int32, filename, contentType string, data []byte) error
	DeleteAttachment(ctx context.Context, idDoc int32) error
	UpdateOrderPayments(ctx context.Context, id int32, rows []models.DocumentPaymentRowBase) error
	AddInvoicePayment(ctx context.Context, id int32, payload models.DocumentPaymentPayload) error
	ListCustomerPayments(ctx context.Context, q models.ListQuery) ([]models.CustomerPayment, error)
	GetCustomerPayment(ctx context.Context, id int32) (*models.CustomerPayment, error)
	CreateCustomerPayment(ctx context.Context, p *models.CustomerPayment) (*models.CustomerPayment, error)
	UpdateCustomerPayment(ctx context.Context, id int32, p *models.CustomerPayment) (*models.CustomerPayment, error)
	DeleteCustomerPayment(ctx context.Context, id int32) error
	GetCustomerPaymentByDoc(ctx context.Context, idDoc int32) (*models.CustomerPayment, error)
	ListOpenBalance(ctx context.Context, idCustomer string, limit, offset int) ([]models.OpenBalanceRow, error)
	SaleReports(ctx context.Context, reportType, startDate, endDate string) (map[string]any, error)
	SubmitEInvoice(ctx context.Context, idDoc int32) (map[string]any, error)
}

type saleDocumentRepository struct {
	db *pgxpool.Pool
}

func NewSaleDocumentRepository(db *pgxpool.Pool) SaleDocumentRepository {
	return &saleDocumentRepository{db: db}
}

const saleDocHeaderSQL = `
	SELECT id, COALESCE(id_document_type,0), COALESCE(id_contact,0), COALESCE(id_customer,''), COALESCE(id_vendor,''),
		doc_date, doc_status, COALESCE(doc_currency,''), tdsc1, tdsc2, tdsc3, tdsc4,
		COALESCE(id_bu,''), COALESCE(id_numerator,0), COALESCE(doc_number,''), COALESCE(sequential_number,0),
		COALESCE(note,''), COALESCE(internal_note,''), id_payment_term, COALESCE(reference,''), COALESCE(dest_id_storage,'')
	FROM sale_documents WHERE company_id = $1 AND deleted = false`

func (r *saleDocumentRepository) ListDocuments(ctx context.Context, kind string, q models.SaleDocumentListQuery) ([]models.Document, error) {
	sql := saleDocHeaderSQL + ` AND doc_kind = $2`
	args := []any{defaultCompanyID, kind}
	n := 3
	if s := strings.TrimSpace(q.FreeText); s != "" {
		sql += fmt.Sprintf(` AND (doc_number ILIKE $%d OR note ILIKE $%d OR id_customer ILIKE $%d)`, n, n, n)
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
		return nil, fmt.Errorf("list sale documents: %w", err)
	}
	defer rows.Close()
	return scanDocumentHeaders(rows)
}

func (r *saleDocumentRepository) GetDocument(ctx context.Context, id int32, kind string, onlyHeader bool) (*models.Document, error) {
	sql := saleDocHeaderSQL + ` AND id = $2`
	args := []any{defaultCompanyID, id}
	if strings.TrimSpace(kind) != "" {
		sql += ` AND doc_kind = $3`
		args = append(args, kind)
	}
	row := r.db.QueryRow(ctx, sql, args...)
	doc, err := scanDocumentHeader(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get sale document: %w", err)
	}
	if onlyHeader {
		return doc, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT id_pos, COALESCE(id_material,''), COALESCE(id_attribute_combination,0), COALESCE(id_pos_type,1),
			description, price, quantity, COALESCE(packages,1), dsc1, dsc2, dsc3, dsc4, COALESCE(id_vat,''), COALESCE(um,'')
		FROM sale_document_rows WHERE document_id = $1 ORDER BY id_pos
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

func (r *saleDocumentRepository) CreateDocument(ctx context.Context, kind string, doc *models.Document) (*models.Document, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var seq int64
	if err := tx.QueryRow(ctx, `SELECT nextval('sale_doc_number_seq')`).Scan(&seq); err != nil {
		return nil, err
	}
	docNumber := fmt.Sprintf("%s-%d", strings.ToUpper(kind[:min(3, len(kind))]), seq)
	docDate := models.MillisToTime(doc.DocDate)

	var id int32
	var idPaymentTerm *int32
	if doc.IDPaymentTerm != nil {
		idPaymentTerm = doc.IDPaymentTerm
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO sale_documents (
			doc_kind, id_document_type, id_contact, id_customer, id_vendor, doc_date, doc_status, doc_currency,
			tdsc1, tdsc2, tdsc3, tdsc4, id_bu, id_numerator, doc_number, sequential_number,
			note, internal_note, id_payment_term, reference, dest_id_storage
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		RETURNING id
	`, kind, nullInt32(doc.IDDocumentType), nullInt32(doc.IDContact), emptyAsNull(doc.IDCustomer), emptyAsNull(doc.IDVendor),
		docDate, defaultStr(doc.DocStatus, "CREATED"), defaultStr(doc.DocCurrency, "EUR"),
		doc.Tdsc1, doc.Tdsc2, doc.Tdsc3, doc.Tdsc4, emptyAsNull(doc.IDBu), nullInt32(doc.IDNumerator),
		docNumber, int32(seq), emptyAsNull(doc.Note), emptyAsNull(doc.InternalNote), idPaymentTerm,
		emptyAsNull(doc.Reference), emptyAsNull(doc.DestIDStorage)).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert sale document: %w", err)
	}
	if err := r.insertRowsAndPayments(ctx, tx, id, doc); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetDocument(ctx, id, "", false)
}

func (r *saleDocumentRepository) UpdateDocument(ctx context.Context, id int32, kind string, doc *models.Document) (*models.Document, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	sql := `
		UPDATE sale_documents SET
			id_contact=$2, id_customer=$3, doc_date=$4, doc_status=$5, doc_currency=$6,
			tdsc1=$7, tdsc2=$8, tdsc3=$9, tdsc4=$10, id_bu=$11, id_numerator=$12,
			note=$13, internal_note=$14, id_payment_term=$15, reference=$16, dest_id_storage=$17, updated_at=now()
		WHERE id=$1 AND company_id=$18 AND deleted=false`
	args := []any{
		id, nullInt32(doc.IDContact), emptyAsNull(doc.IDCustomer), models.MillisToTime(doc.DocDate),
		defaultStr(doc.DocStatus, "CREATED"), defaultStr(doc.DocCurrency, "EUR"),
		doc.Tdsc1, doc.Tdsc2, doc.Tdsc3, doc.Tdsc4, emptyAsNull(doc.IDBu), nullInt32(doc.IDNumerator),
		emptyAsNull(doc.Note), emptyAsNull(doc.InternalNote), doc.IDPaymentTerm,
		emptyAsNull(doc.Reference), emptyAsNull(doc.DestIDStorage), defaultCompanyID,
	}
	if strings.TrimSpace(kind) != "" {
		sql += ` AND doc_kind = $19`
		args = append(args, kind)
	}
	tag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("sale document not found")
	}
	if _, err := tx.Exec(ctx, `DELETE FROM sale_document_rows WHERE document_id=$1`, id); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM sale_document_payments WHERE document_id=$1`, id); err != nil {
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

func (r *saleDocumentRepository) DeleteDocument(ctx context.Context, id int32, kind string) error {
	sql := `UPDATE sale_documents SET deleted=true, updated_at=now() WHERE id=$1 AND company_id=$2 AND deleted=false`
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
		return fmt.Errorf("sale document not found")
	}
	return nil
}

func (r *saleDocumentRepository) TransformDocument(ctx context.Context, sourceID int32, targetKind string, docDate *int64, docNumber string) (*models.Document, error) {
	src, err := r.GetDocument(ctx, sourceID, "", false)
	if err != nil || src == nil {
		return nil, fmt.Errorf("source document not found")
	}
	if docDate != nil {
		src.DocDate = docDate
	}
	src.DocStatus = "CREATED"
	src.ID = 0
	out, err := r.CreateDocument(ctx, targetKind, src)
	if err != nil {
		return nil, err
	}
	_, _ = r.db.Exec(ctx, `UPDATE sale_documents SET source_document_id=$2 WHERE id=$1`, out.ID, sourceID)
	if docNumber != "" {
		_, _ = r.db.Exec(ctx, `UPDATE sale_documents SET doc_number=$2 WHERE id=$1`, out.ID, docNumber)
	}
	return r.GetDocument(ctx, out.ID, "", false)
}

func (r *saleDocumentRepository) TransformDocumentList(ctx context.Context, targetKind string, payload models.DocumentListTransformation) ([]models.Document, error) {
	ids := strings.Split(payload.IDDocumentList, ",")
	var out []models.Document
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

func (r *saleDocumentRepository) GetPDF(ctx context.Context, id int32) ([]byte, error) {
	var pdf []byte
	err := r.db.QueryRow(ctx, `SELECT pdf_data FROM sale_documents WHERE id=$1 AND deleted=false`, id).Scan(&pdf)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if len(pdf) > 0 {
		return pdf, nil
	}
	// Minimal PDF stub
	return []byte("%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n"), nil
}

func (r *saleDocumentRepository) GetAttachment(ctx context.Context, idDoc int32) (*models.SaleDocumentAttachment, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, filename, COALESCE(content_type,'') FROM sale_document_attachments
		WHERE document_id=$1 ORDER BY id DESC LIMIT 1
	`, idDoc)
	var a models.SaleDocumentAttachment
	if err := row.Scan(&a.ID, &a.Filename, &a.ContentType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *saleDocumentRepository) SaveAttachment(ctx context.Context, idDoc int32, filename, contentType string, data []byte) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO sale_document_attachments (document_id, filename, content_type, data) VALUES ($1,$2,$3,$4)
	`, idDoc, filename, contentType, data)
	return err
}

func (r *saleDocumentRepository) DeleteAttachment(ctx context.Context, idDoc int32) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sale_document_attachments WHERE document_id=$1`, idDoc)
	return err
}

func (r *saleDocumentRepository) UpdateOrderPayments(ctx context.Context, id int32, rows []models.DocumentPaymentRowBase) error {
	for _, row := range rows {
		_, err := r.db.Exec(ctx, `
			UPDATE sale_document_payments SET paid_amount=$3
			WHERE document_id=$1 AND rate_number=$2
		`, id, row.RateNumber, row.PaidAmount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *saleDocumentRepository) AddInvoicePayment(ctx context.Context, id int32, payload models.DocumentPaymentPayload) error {
	_, err := r.db.Exec(ctx, `
		UPDATE sale_document_payments SET paid_amount = paid_amount + $2, payment_date = now()
		WHERE document_id=$1 AND rate_number = COALESCE((SELECT MIN(rate_number) FROM sale_document_payments WHERE document_id=$1), 1)
	`, id, payload.Amount)
	return err
}

func (r *saleDocumentRepository) ListCustomerPayments(ctx context.Context, q models.ListQuery) ([]models.CustomerPayment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, COALESCE(id_document_type,0), COALESCE(reference,''), id_agent,
			total_paid_ic, total_invoice_ic, total_remaining_ic, doc_date, COALESCE(note,''), COALESCE(doc_number,''), deleted
		FROM customer_payments WHERE company_id=$1 AND deleted=false
		ORDER BY id DESC LIMIT $2 OFFSET $3
	`, defaultCompanyID, limitOrDefault(q.Limit), q.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCustomerPayments(rows)
}

func (r *saleDocumentRepository) GetCustomerPayment(ctx context.Context, id int32) (*models.CustomerPayment, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(id_document_type,0), COALESCE(reference,''), id_agent,
			total_paid_ic, total_invoice_ic, total_remaining_ic, doc_date, COALESCE(note,''), COALESCE(doc_number,''), deleted
		FROM customer_payments WHERE id=$1 AND company_id=$2 AND deleted=false
	`, id, defaultCompanyID)
	return scanCustomerPaymentRow(row)
}

func (r *saleDocumentRepository) CreateCustomerPayment(ctx context.Context, p *models.CustomerPayment) (*models.CustomerPayment, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO customer_payments (id_document_type, reference, id_agent, total_paid_ic, total_invoice_ic, total_remaining_ic, doc_date, note, doc_number)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id
	`, p.IDDocumentType, emptyAsNull(p.Reference), p.IDAgent, p.TotalPaidIc, p.TotalInvoiceIc, p.TotalRemainingIc,
		models.MillisToTime(p.DocDate), emptyAsNull(p.Note), emptyAsNull(p.DocNumber))
	var id int32
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return r.GetCustomerPayment(ctx, id)
}

func (r *saleDocumentRepository) UpdateCustomerPayment(ctx context.Context, id int32, p *models.CustomerPayment) (*models.CustomerPayment, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE customer_payments SET reference=$2, total_paid_ic=$3, total_invoice_ic=$4, total_remaining_ic=$5,
			doc_date=$6, note=$7, updated_at=now()
		WHERE id=$1 AND company_id=$8 AND deleted=false
	`, id, emptyAsNull(p.Reference), p.TotalPaidIc, p.TotalInvoiceIc, p.TotalRemainingIc,
		models.MillisToTime(p.DocDate), emptyAsNull(p.Note), defaultCompanyID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("customer payment not found")
	}
	return r.GetCustomerPayment(ctx, id)
}

func (r *saleDocumentRepository) DeleteCustomerPayment(ctx context.Context, id int32) error {
	tag, err := r.db.Exec(ctx, `UPDATE customer_payments SET deleted=true WHERE id=$1 AND company_id=$2`, id, defaultCompanyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("customer payment not found")
	}
	return nil
}

func (r *saleDocumentRepository) GetCustomerPaymentByDoc(ctx context.Context, idDoc int32) (*models.CustomerPayment, error) {
	// Link via reference doc id as string stub
	row := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(id_document_type,0), COALESCE(reference,''), id_agent,
			total_paid_ic, total_invoice_ic, total_remaining_ic, doc_date, COALESCE(note,''), COALESCE(doc_number,''), deleted
		FROM customer_payments WHERE reference=$1 AND company_id=$2 AND deleted=false LIMIT 1
	`, strconv.Itoa(int(idDoc)), defaultCompanyID)
	return scanCustomerPaymentRow(row)
}

func (r *saleDocumentRepository) ListOpenBalance(ctx context.Context, idCustomer string, limit, offset int) ([]models.OpenBalanceRow, error) {
	sql := `
		SELECT COALESCE(d.id_customer,''), d.id, COALESCE(d.doc_number,''),
			COALESCE(SUM(p.amount),0), COALESCE(SUM(p.paid_amount),0),
			COALESCE(SUM(p.amount),0) - COALESCE(SUM(p.paid_amount),0),
			CASE WHEN COALESCE(SUM(p.amount),0) <= COALESCE(SUM(p.paid_amount),0) THEN 'PAID_STATUS' ELSE 'PENDING_STATUS' END
		FROM sale_documents d
		LEFT JOIN sale_document_payments p ON p.document_id = d.id
		WHERE d.doc_kind = 'invoice' AND d.deleted = false AND d.company_id = $1`
	args := []any{defaultCompanyID}
	n := 2
	if s := strings.TrimSpace(idCustomer); s != "" {
		sql += fmt.Sprintf(` AND d.id_customer = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += ` GROUP BY d.id, d.id_customer, d.doc_number`
	sql += fmt.Sprintf(` ORDER BY d.id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(limit), offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.OpenBalanceRow
	out = make([]models.OpenBalanceRow, 0)
	for rows.Next() {
		var row models.OpenBalanceRow
		if err := rows.Scan(&row.IDCustomer, &row.IDDocument, &row.DocNumber, &row.Amount, &row.PaidAmount, &row.Remaining, &row.PaymentStatus); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *saleDocumentRepository) SaleReports(ctx context.Context, reportType, startDate, endDate string) (map[string]any, error) {
	var revenue float64
	_ = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(r.price * r.quantity),0)
		FROM sale_document_rows r
		JOIN sale_documents d ON d.id = r.document_id
		WHERE d.doc_kind IN ('invoice','ticket') AND d.deleted = false
	`).Scan(&revenue)
	openOrders := 0
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM sale_documents WHERE doc_kind='order' AND deleted=false AND doc_status <> 'COMPLETED_ORDER'`).Scan(&openOrders)
	return map[string]any{
		"reportType":      defaultStr(reportType, "ALL"),
		"revenues":        revenue,
		"openSalesOrders": openOrders,
	}, nil
}

func (r *saleDocumentRepository) SubmitEInvoice(ctx context.Context, idDoc int32) (map[string]any, error) {
	doc, err := r.GetDocument(ctx, idDoc, models.SaleDocInvoice, true)
	if err != nil || doc == nil {
		return nil, fmt.Errorf("invoice not found")
	}
	return map[string]any{"idDoc": idDoc, "status": "QUEUED", "message": "e-invoice submission queued"}, nil
}

func (r *saleDocumentRepository) insertRowsAndPayments(ctx context.Context, tx pgx.Tx, docID int32, doc *models.Document) error {
	for i, row := range doc.Rows {
		pos := row.IDPos
		if pos == 0 {
			pos = int32(i + 1)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO sale_document_rows (document_id, id_pos, id_material, id_attribute_combination, id_pos_type,
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
			INSERT INTO sale_document_payments (document_id, rate_number, description, amount, paid_amount, due_date, payment_date)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, docID, rate, emptyAsNull(p.Description), p.Amount, p.PaidAmount,
			millisToTimePtr(p.Date), millisToTimePtr(p.PaymentDate)); err != nil {
			return err
		}
	}
	return nil
}

func (r *saleDocumentRepository) loadPayments(ctx context.Context, docID int32) ([]models.DocumentPaymentRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT rate_number, COALESCE(description,''), amount, paid_amount, due_date, payment_date
		FROM sale_document_payments WHERE document_id=$1 ORDER BY rate_number
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

func scanDocumentHeaders(rows pgx.Rows) ([]models.Document, error) {
	var out []models.Document
	for rows.Next() {
		d, err := scanDocumentHeader(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

func scanDocumentHeader(s interface{ Scan(...any) error }) (*models.Document, error) {
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

func scanCustomerPayments(rows pgx.Rows) ([]models.CustomerPayment, error) {
	var out []models.CustomerPayment
	for rows.Next() {
		p, err := scanCustomerPaymentFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func scanCustomerPaymentRow(row pgx.Row) (*models.CustomerPayment, error) {
	p, err := scanCustomerPaymentFromScanner(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func scanCustomerPaymentFromScanner(s interface{ Scan(...any) error }) (*models.CustomerPayment, error) {
	var p models.CustomerPayment
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

func nullInt32(v int32) any {
	if v == 0 {
		return nil
	}
	return v
}

func defaultStr(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

func defaultFloat(v, def float64) float64 {
	if v == 0 {
		return def
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
