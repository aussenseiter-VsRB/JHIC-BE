ALTER TABLE users ADD COLUMN class TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN jurusan TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN position TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX idx_users_wali_kelas ON users (class) WHERE position = 'wali_kelas' AND class <> '';
CREATE UNIQUE INDEX idx_users_kaprog ON users (jurusan) WHERE position = 'kaprog' AND jurusan <> '';
CREATE UNIQUE INDEX idx_users_bk ON users (position) WHERE position = 'bk';
CREATE UNIQUE INDEX idx_users_kesiswaan ON users (position) WHERE position = 'kesiswaan';
