CREATE TABLE contacts (
  id BIGSERIAL PRIMARY KEY,
  role integer DEFAULT 1,
  title character varying,
  name character varying,
  phone character varying,
  email character varying,
  start_at timestamp(0) without time zone,
  sign character varying,
  details jsonb DEFAULT '{}'::jsonb,
  company_id bigint REFERENCES companies (id) ON DELETE CASCADE,
  user_id bigint REFERENCES users (id) ON DELETE SET NULL,
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);