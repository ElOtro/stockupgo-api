DROP TABLE IF EXISTS invoices CASCADE;
DROP INDEX IF EXISTS invoices_date_index;
DROP INDEX IF EXISTS invoices_organisation_id_index;
DROP INDEX IF EXISTS invoices_bank_account_id_index;
DROP INDEX IF EXISTS invoices_company_id_index;
DROP INDEX IF EXISTS invoices_agreement_id_index;
DROP INDEX IF EXISTS invoices_project_id_index;
DROP INDEX IF EXISTS invoices_uuid_index;
DROP INDEX IF EXISTS invoices_search_vector_index;
DROP FUNCTION IF EXISTS full_search_for_invoices();