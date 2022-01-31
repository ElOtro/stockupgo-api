CREATE TABLE products (
  id BIGSERIAL PRIMARY KEY,
  is_active boolean DEFAULT true,
  product_type integer,
  name character varying,
  description text,
  sku character varying(25),
  price numeric(15,2) DEFAULT 0.0,
  vat_rate_id bigint REFERENCES vat_rates (id),
  unit_id bigint REFERENCES units (id),
  user_id bigint REFERENCES users (id) ON DELETE SET NULL,
  search_vector tsvector,
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);