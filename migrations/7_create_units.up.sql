CREATE TABLE units (
  id BIGSERIAL PRIMARY KEY,
  code character varying(3),
  name character varying(25),
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);
