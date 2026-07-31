package pkl

import (
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
)

const (
	StatusPending     = "pending"
	StatusAccepted    = "accepted"
	StatusRejected    = "rejected"
	StatusNeedsAction = "needs_further_action"
	StatusCancelled   = "cancelled"
)

const (
	PositionWaliKelas = "wali_kelas"
	PositionBK        = "bk"
	PositionKesiswaan = "kesiswaan"
	PositionKaprog    = "kaprog"
)

const (
	StepPending     = "pending"
	StepApproved    = "approved"
	StepRejected    = "rejected"
	StepNeedsAction = "needs_further_action"
)

const DecisionApprove = "approve"
const DecisionReject = "reject"
const DecisionNeedsAction = "needs_further_action"

var ApprovalOrder = []string{PositionWaliKelas, PositionBK, PositionKesiswaan, PositionKaprog}

type PklRequest struct {
	ID           id.ID     `json:"id"`
	RequesterID  id.ID     `json:"requester_id"`
	Company      string    `json:"company"`
	Location     string    `json:"location"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	CancelReason string    `json:"cancel_reason,omitempty"`
	CurrentStep  int       `json:"current_step"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Steps        []Step    `json:"steps,omitempty"`
}

type Step struct {
	ID         id.ID      `json:"id"`
	RequestID  id.ID      `json:"request_id"`
	Position   string     `json:"position"`
	ApproverID id.ID      `json:"approver_id"`
	Status     string     `json:"status"`
	Note       string     `json:"note,omitempty"`
	Sequence   int        `json:"sequence"`
	DecidedAt  *time.Time `json:"decided_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
