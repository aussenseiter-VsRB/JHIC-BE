DROP INDEX IF EXISTS idx_users_wali_kelas;
DROP INDEX IF EXISTS idx_users_kaprog;
DROP INDEX IF EXISTS idx_users_bk;
DROP INDEX IF EXISTS idx_users_kesiswaan;
ALTER TABLE users DROP COLUMN IF EXISTS class, DROP COLUMN IF EXISTS jurusan, DROP COLUMN IF EXISTS position;
