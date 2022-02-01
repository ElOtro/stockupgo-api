CREATE TABLE invoices (
  id BIGSERIAL PRIMARY KEY,
  is_active boolean DEFAULT false,
  date timestamp without time zone,
  number character varying(11),
  organisation_id bigint REFERENCES organisations (id) ON DELETE CASCADE,
  bank_account_id bigint REFERENCES bank_accounts (id),
  company_id bigint REFERENCES companies (id) ON DELETE CASCADE,
  agreement_id bigint REFERENCES agreements (id),
  project_id bigint REFERENCES projects (id),
  amount numeric(15,2) DEFAULT 0.0,
  discount numeric(15,2) DEFAULT 0.0,
  vat numeric(15,2) DEFAULT 0.0,
  user_id bigint REFERENCES users (id),
  search_vector tsvector,
  uuid uuid DEFAULT uuid_generate_v4(),
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) without time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS invoices_date_index ON invoices USING btree (date);
CREATE INDEX IF NOT EXISTS invoices_organisation_id_index ON invoices USING btree (organisation_id);
CREATE INDEX IF NOT EXISTS invoices_bank_account_id_index ON invoices USING btree (bank_account_id);
CREATE INDEX IF NOT EXISTS invoices_company_id_index ON invoices USING btree (company_id);
CREATE INDEX IF NOT EXISTS invoices_agreement_id_index ON invoices USING btree (agreement_id);
CREATE INDEX IF NOT EXISTS invoices_project_id_index ON invoices USING btree (project_id);
CREATE INDEX IF NOT EXISTS invoices_uuid_index ON invoices USING btree (uuid);
CREATE INDEX IF NOT EXISTS invoices_search_vector_index ON invoices USING GIN (search_vector);


CREATE OR REPLACE FUNCTION full_search_for_invoices() RETURNS trigger AS $$
declare
  companies record;
  agreements record;
begin
  select name into companies from companies where (id = new.company_id);
  select name into agreements from agreements where (id = new.agreement_id);

  new.search_vector :=
      setweight(to_tsvector('pg_catalog.simple', coalesce(new.number, '')), 'A') ||
      setweight(to_tsvector('pg_catalog.simple', coalesce(companies.name, '')), 'B') ||
      setweight(to_tsvector('pg_catalog.simple', coalesce(replace(agreements.name, '-', ' '), '')), 'B');
    return new;
  end
$$ LANGUAGE plpgsql;

CREATE TRIGGER invoice_search BEFORE INSERT OR UPDATE 
ON invoices 
FOR EACH ROW EXECUTE PROCEDURE full_search_for_invoices();
