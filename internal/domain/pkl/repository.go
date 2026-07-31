package pkl

import "context"

type Repository interface {
	CreateRequest(ctx context.Context, req *PklRequest, steps []Step) error
	ByID(ctx context.Context, id string) (*PklRequest, error)
	ListByRequester(ctx context.Context, requesterID string) ([]PklRequest, error)
	ListForApprover(ctx context.Context, approverID string) ([]PklRequest, error)
	ListAll(ctx context.Context) ([]PklRequest, error)
	StepsByRequest(ctx context.Context, requestID string) ([]Step, error)
	Decide(ctx context.Context, req *PklRequest, step *Step, expectedReqStatus, expectedStepStatus string) error
	Cancel(ctx context.Context, req *PklRequest) error
}
