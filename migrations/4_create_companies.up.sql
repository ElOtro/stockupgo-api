CREATE TABLE companies (
  id BIGSERIAL PRIMARY KEY,
  logo character varying,
  name character varying(100),
  full_name character varying(250),
  company_type integer DEFAULT 1,
  details jsonb DEFAULT '{}'::jsonb,
  user_id bigint REFERENCES users (id) ON DELETE SET NULL,
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);
CREATE INDEX companies_user_id_index ON companies USING btree (user_id);
CREATE INDEX companies_destroyed_at_index ON companies USING btree (destroyed_at);
