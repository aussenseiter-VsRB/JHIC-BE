CREATE TABLE IF NOT EXISTS spmb_registrations (
    id BIGINT PRIMARY KEY,
    nama TEXT NOT NULL,
    nik TEXT NOT NULL,
    nisn TEXT NOT NULL DEFAULT '',
    kk_no TEXT NOT NULL DEFAULT '',
    tempat_lahir TEXT NOT NULL DEFAULT '',
    tanggal_lahir TEXT NOT NULL DEFAULT '',
    jenis_kelamin TEXT NOT NULL DEFAULT '',
    agama TEXT NOT NULL DEFAULT '',
    alamat TEXT NOT NULL DEFAULT '',
    asal_sekolah TEXT NOT NULL DEFAULT '',
    no_hp TEXT NOT NULL DEFAULT '',
    nama_ayah TEXT NOT NULL DEFAULT '',
    nama_ibu TEXT NOT NULL DEFAULT '',
    jurusan TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'proses',
    cancel_reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT spmb_registrations_status_check CHECK (status IN ('proses', 'approve', 'cancel'))
);

CREATE INDEX idx_spmb_registrations_status ON spmb_registrations (status);
CREATE INDEX idx_spmb_registrations_jurusan ON spmb_registrations (jurusan);
