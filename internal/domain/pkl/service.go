package pkl

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
)

type Users interface {
	ByID(ctx context.Context, id string) (*user.User, error)
	FindByPosition(ctx context.Context, position, class, jurusan string) (*user.User, error)
}

type Service struct {
	repo  Repository
	users Users
}

func NewService(repo Repository, users Users) *Service {
	return &Service{repo: repo, users: users}
}

func (s *Service) Create(ctx context.Context, requesterID, company, location string, startDate, endDate time.Time, description string) (*PklRequest, error) {
	requester, err := s.users.ByID(ctx, requesterID)
	if err != nil {
		return nil, err
	}
	if requester == nil {
		return nil, fmt.Errorf("requester not found")
	}
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end date must be on or after start date")
	}

	steps := make([]Step, 0, len(ApprovalOrder))
	for i, position := range ApprovalOrder {
		class, jurusan := "", ""
		switch position {
		case PositionWaliKelas:
			class = requester.Class
		case PositionKaprog:
			jurusan = requester.Jurusan
		}
		approver, err := s.users.FindByPosition(ctx, position, class, jurusan)
		if err != nil {
			return nil, err
		}
		if approver == nil {
			return nil, fmt.Errorf("no %s assigned for this request", position)
		}
		now := time.Now().UTC()
		steps = append(steps, Step{
			ID:         id.New(),
			Position:   position,
			ApproverID: approver.ID,
			Status:     StepPending,
			Sequence:   i + 1,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}

	now := time.Now().UTC()
	reqID := id.New()
	for i := range steps {
		steps[i].RequestID = reqID
	}
	req := &PklRequest{
		ID:          reqID,
		RequesterID: requesterID,
		Company:     company,
		Location:    location,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: description,
		Status:      StatusPending,
		CurrentStep: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
		Steps:       steps,
	}
	if err := s.repo.CreateRequest(ctx, req, steps); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Service) List(ctx context.Context, callerID, role string) ([]PklRequest, error) {
	var list []PklRequest
	var err error
	switch role {
	case "admin":
		list, err = s.repo.ListAll(ctx)
	case "guru":
		list, err = s.repo.ListForApprover(ctx, callerID)
	default:
		list, err = s.repo.ListByRequester(ctx, callerID)
	}
	if err != nil {
		return nil, err
	}
	for i := range list {
		steps, err := s.repo.StepsByRequest(ctx, list[i].ID)
		if err != nil {
			return nil, err
		}
		list[i].Steps = steps
	}
	return list, nil
}

func (s *Service) Get(ctx context.Context, callerID, role, requestID string) (*PklRequest, error) {
	req, err := s.repo.ByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, fmt.Errorf("pkl request not found")
	}
	steps, err := s.repo.StepsByRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	req.Steps = steps

	switch role {
	case "admin":
	case "guru":
		if !isApprover(callerID, steps) {
			return nil, fmt.Errorf("forbidden: not an approver on this request")
		}
	default:
		if req.RequesterID != callerID {
			return nil, fmt.Errorf("forbidden: not the requester")
		}
	}
	return req, nil
}

func (s *Service) Decide(ctx context.Context, callerID, requestID, decision, note string) (*PklRequest, error) {
	req, err := s.repo.ByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, fmt.Errorf("pkl request not found")
	}
	steps, err := s.repo.StepsByRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("pkl request has no approval steps")
	}

	expectedReqStatus := req.Status
	expectedStepStatus := ""
	var myStep *Step
	switch req.Status {
	case StatusPending:
		if req.CurrentStep < 1 || req.CurrentStep > len(steps) {
			return nil, fmt.Errorf("invalid request step state")
		}
		step := &steps[req.CurrentStep-1]
		if step.ApproverID != callerID {
			return nil, fmt.Errorf("forbidden: not your step to decide")
		}
		if step.Status != StepPending {
			return nil, fmt.Errorf("step already decided")
		}
		myStep = step
		expectedStepStatus = StepPending
	case StatusNeedsAction:
		for i := range steps {
			if steps[i].ApproverID == callerID && steps[i].Status == StepNeedsAction {
				myStep = &steps[i]
				break
			}
		}
		if myStep == nil {
			return nil, fmt.Errorf("forbidden: not your step to decide")
		}
		expectedStepStatus = StepNeedsAction
	default:
		return nil, fmt.Errorf("request is not awaiting a decision")
	}

	switch decision {
	case DecisionApprove:
		myStep.Status = StepApproved
		if myStep.Sequence == len(steps) {
			req.Status = StatusAccepted
		} else {
			req.Status = StatusPending
			req.CurrentStep++
		}
	case DecisionReject:
		myStep.Status = StepRejected
		req.Status = StatusRejected
	case DecisionNeedsAction:
		if req.Status == StatusNeedsAction {
			return nil, fmt.Errorf("request already needs further action")
		}
		myStep.Status = StepNeedsAction
		req.Status = StatusNeedsAction
	default:
		return nil, fmt.Errorf("invalid decision: must be approve, reject or needs_further_action")
	}

	myStep.Note = note
	now := time.Now().UTC()
	myStep.DecidedAt = &now
	myStep.UpdatedAt = now
	req.UpdatedAt = now

	if err := s.repo.Decide(ctx, req, myStep, expectedReqStatus, expectedStepStatus); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Service) Cancel(ctx context.Context, callerID, requestID, reason string) (*PklRequest, error) {
	if reason == "" {
		return nil, fmt.Errorf("cancellation reason is required")
	}
	req, err := s.repo.ByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, fmt.Errorf("pkl request not found")
	}
	if req.RequesterID != callerID {
		return nil, fmt.Errorf("forbidden: not the requester")
	}
	if req.Status != StatusPending && req.Status != StatusNeedsAction {
		return nil, fmt.Errorf("only pending or needs_further_action requests can be cancelled")
	}
	req.Status = StatusCancelled
	req.CancelReason = reason
	req.UpdatedAt = time.Now().UTC()
	if err := s.repo.Cancel(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func isApprover(userID string, steps []Step) bool {
	return slices.ContainsFunc(steps, func(s Step) bool { return s.ApproverID == userID })
}
