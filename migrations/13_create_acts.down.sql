DROP TABLE IF EXISTS acts CASCADE;
DROP INDEX IF EXISTS acts_date_index;
DROP INDEX IF EXISTS acts_uuid_index;
DROP INDEX IF EXISTS acts_search_vector_index;
DROP FUNCTION IF EXISTS full_search_for_acts();