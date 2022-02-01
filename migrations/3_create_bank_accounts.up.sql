CREATE TABLE bank_accounts (
  id BIGSERIAL PRIMARY KEY,
  organisation_id bigint REFERENCES organisations (id) ON DELETE CASCADE,
  is_default boolean DEFAULT false,
  account_type integer DEFAULT 1,
  name character varying,
  details jsonb DEFAULT '{}'::jsonb,
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) without time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);
CREATE INDEX bank_accounts_organisation_id_index ON bank_accounts USING btree (organisation_id);
CREATE INDEX bank_accounts_destroyed_at_index ON bank_accounts USING btree (destroyed_at);