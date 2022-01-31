CREATE TABLE projects (
  id BIGSERIAL PRIMARY KEY,
  organisation_id bigint REFERENCES organisations (id) ON DELETE CASCADE,
  name character varying(255),
  uuid uuid DEFAULT uuid_generate_v4(),
  destroyed_at timestamp(0) without time zone,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) without time zone NOT NULL DEFAULT NOW()
);
