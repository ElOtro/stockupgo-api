CREATE TABLE agreements (
  id BIGSERIAL PRIMARY KEY,
  start_at date,
  end_at date,
  name character varying,
  company_id bigint REFERENCES companies (id) ON DELETE CASCADE,
  user_id bigint REFERENCES users (id) ON DELETE SET NULL,
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);
CREATE INDEX agreements_company_id_index ON agreements USING btree (company_id);
CREATE INDEX agreements_user_id_index ON agreements USING btree (user_id);
CREATE INDEX agreements_destroyed_at_index ON agreements USING btree (destroyed_at);