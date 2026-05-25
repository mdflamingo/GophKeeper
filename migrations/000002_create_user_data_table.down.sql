DROP INDEX IF EXISTS idx_user_data_active;
DROP INDEX IF EXISTS idx_user_data_metadata_gin;
DROP INDEX IF EXISTS idx_user_data_deleted_at;
DROP INDEX IF EXISTS idx_user_data_data_type;
DROP INDEX IF EXISTS idx_user_data_user_id;

DROP TABLE IF EXISTS user_data;
DROP TYPE IF EXISTS dtype;