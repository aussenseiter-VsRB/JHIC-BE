package pkl

import (
	"context"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
)

type Repository interface {
	CreateRequest(ctx context.Context, req *PklRequest, steps []Step) error
	ByID(ctx context.Context, id id.ID) (*PklRequest, error)
	ListByRequester(ctx context.Context, requesterID id.ID) ([]PklRequest, error)
	ListForApprover(ctx context.Context, approverID id.ID) ([]PklRequest, error)
	ListAll(ctx context.Context) ([]PklRequest, error)
	StepsByRequest(ctx context.Context, requestID id.ID) ([]Step, error)
	Decide(ctx context.Context, req *PklRequest, step *Step, expectedReqStatus, expectedStepStatus string) error
	Cancel(ctx context.Context, req *PklRequest) error
}
