CREATE TABLE IF NOT EXISTS pkl_requests (
    id BIGINT PRIMARY KEY,
    requester_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company TEXT NOT NULL,
    location TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    description TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    cancel_reason TEXT NOT NULL DEFAULT '',
    current_step INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pkl_requests_status_check CHECK (status IN ('pending', 'accepted', 'rejected', 'needs_further_action', 'cancelled')),
    CONSTRAINT pkl_requests_step_range CHECK (current_step BETWEEN 1 AND 4)
);

CREATE INDEX idx_pkl_requests_requester ON pkl_requests (requester_id);
CREATE INDEX idx_pkl_requests_status ON pkl_requests (status);

CREATE TABLE IF NOT EXISTS pkl_approval_steps (
    id BIGINT PRIMARY KEY,
    pkl_request_id BIGINT NOT NULL REFERENCES pkl_requests(id) ON DELETE CASCADE,
    position TEXT NOT NULL,
    approver_id BIGINT NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'pending',
    note TEXT NOT NULL DEFAULT '',
    sequence INT NOT NULL,
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pkl_approval_steps_position_check CHECK (position IN ('wali_kelas', 'bk', 'kesiswaan', 'kaprog')),
    CONSTRAINT pkl_approval_steps_status_check CHECK (status IN ('pending', 'approved', 'rejected', 'needs_further_action')),
    UNIQUE (pkl_request_id, sequence),
    UNIQUE (pkl_request_id, position)
);

CREATE INDEX idx_pkl_approval_steps_approver ON pkl_approval_steps (approver_id);
