CREATE TABLE vat_rates (
  id BIGSERIAL PRIMARY KEY,
  is_active boolean DEFAULT true,
  is_default boolean DEFAULT false,
  rate numeric(8,2),
  name character varying(50),
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);
CREATE INDEX vat_rates_destroyed_at_index ON vat_rates USING btree (destroyed_at);