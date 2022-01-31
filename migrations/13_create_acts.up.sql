CREATE TABLE acts (
  id BIGSERIAL PRIMARY KEY,
  is_active boolean DEFAULT false,
  date timestamp without time zone,
  number character varying(11),
  organisation_id bigint REFERENCES organisations (id) ON DELETE CASCADE,
  company_id bigint REFERENCES companies (id) ON DELETE CASCADE,
  agreement_id bigint REFERENCES agreements (id),
  project_id bigint REFERENCES projects (id),
  amount numeric(15,2) DEFAULT 0.0,
  vat numeric(15,2) DEFAULT 0.0,
  user_id bigint REFERENCES users (id),
  search_vector tsvector,
  uuid uuid DEFAULT uuid_generate_v4(),
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) without time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS acts_date_index ON acts USING btree (date);
CREATE INDEX IF NOT EXISTS acts_uuid_index ON acts USING btree (uuid);
CREATE INDEX IF NOT EXISTS acts_search_vector_index ON acts USING GIN (search_vector);

CREATE OR REPLACE FUNCTION full_search_for_acts() RETURNS trigger AS $$
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

CREATE TRIGGER act_search BEFORE INSERT OR UPDATE 
ON acts 
FOR EACH ROW EXECUTE PROCEDURE full_search_for_acts();