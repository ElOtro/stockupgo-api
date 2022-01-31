DROP TABLE IF EXISTS invoices CASCADE;
DROP INDEX IF EXISTS invoices_date_index;
DROP INDEX IF EXISTS invoices_uuid_index;
DROP INDEX IF EXISTS invoices_search_vector_index;
DROP FUNCTION IF EXISTS full_search_for_invoices();