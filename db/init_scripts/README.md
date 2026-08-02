# Init scripts — mapare domenii Swagger

Fișierele din acest folder rulează **în ordine alfabetică** la primul start Postgres (`docker-entrypoint-initdb.d`). Prefixul numeric (`010_`, `020_`, …) respectă dependențele FK.

## Implementate

| Swagger tag | Fișier SQL |
|-------------|------------|
| *(extensions)* | `000_extensions.sql` |
| User | `010_user.sql` |
| Country | `020_country.sql` |
| Currency | `030_currency.sql` |
| City | `040_city.sql` |
| Um | `050_um.sql` |
| OfficeType | `060_officetype.sql` |
| ContactRole | `070_contactrole.sql` |
| PaymentTerm | `080_paymentterm.sql` |
| Company | `090_company.sql` |
| CompanySettings | `100_companysettings.sql` |
| Contact | `110_contact.sql` |
| Customer | `120_customer.sql` |
| Vendor | `130_vendor.sql` |
| ProductGroup | `140_productgroup.sql` |
| Product | `150_product.sql` |
| Attribute | `160_attribute.sql` |
| Pricelist | `170_pricelist.sql` |
| Storage | `180_storage.sql` |
| StorageLocation | `190_storagelocation.sql` |
| Lot | `200_lot.sql` |
| Stock | `210_stock.sql` |
| MachineDataTracking | `220_machinedatatracking.sql` |
| SaleDocument | `230_saledocument.sql` |
| PurchaseDocument | `240_purchasedocument.sql` |
| Accounting | `250_accounting.sql` |
| Crm | `260_crm.sql` |
| Calendar | `270_calendar.sql` |
| PersonalActivity | `280_personalactivity.sql` |
| User | `290_user_management.sql` |
| Settings | `300_settings.sql` |

## Viitoare (Task 16+)

La implementare, adaugă fișier nou cu prefix după ultimul existent:

| Swagger tag | Task |
|-------------|------|
| LoggedUser | 3 (revizuit) |
| Crm | — |
| PuchaseDocument | — |
| Accounting | — |
| Calendar | — |
| Ecommerce | — |
| Email | — |
| GlobalNotification | — |
| Message / MessageGroups | — |
| PersonalActivity | — |
| Plugin | — |
| RepoFile / RepoMedia | 16 |
| Settings | — |
| Social | — |
| Task | 12 (parțial) |
| Tilby Sales | — |

## Reset după modificări schema

```bash
docker compose down -v
docker compose up -d --build
```
