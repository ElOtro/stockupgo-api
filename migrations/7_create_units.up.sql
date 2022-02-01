CREATE TABLE units (
  id BIGSERIAL PRIMARY KEY,
  name character varying(25),
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);
CREATE INDEX units_destroyed_at_index ON units USING btree (destroyed_at);
