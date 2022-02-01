CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE organisations (
  id BIGSERIAL PRIMARY KEY,
  name character varying(50),
  full_name character varying(255),
  ceo character varying(150),
  ceo_title character varying(40),
  cfo character varying(150),
  cfo_title character varying(40),
  stamp character varying,
  ceo_sign character varying,
  cfo_sign character varying,
  is_vat_payer boolean DEFAULT false,
  details jsonb DEFAULT '{}'::jsonb,
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()  
);

CREATE INDEX organisations_destroyed_at_index ON organisations USING btree (destroyed_at);