CREATE TABLE act_items (
  id BIGSERIAL PRIMARY KEY,
  act_id bigint REFERENCES acts (id) ON DELETE CASCADE,
  position integer DEFAULT 0,
  product_id bigint REFERENCES products (id),
  description character varying(1024),
  unit_id bigint REFERENCES units (id),
  quantity numeric(8,3) DEFAULT 0,
  price numeric(15,2) DEFAULT 0.0,
  amount numeric(15,2) DEFAULT 0.0,
  vat_rate_id bigint REFERENCES vat_rates (id),
  vat numeric(15,2) DEFAULT 0.0,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);