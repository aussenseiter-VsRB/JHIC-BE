package pkl_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/pkl"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/pkl/mocks"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func approver(position string, id id.ID) *user.User {
	return &user.User{ID: id, Role: "guru", Position: position}
}

func fullSteps(requestID id.ID) []pkl.Step {
	now := time.Now().UTC()
	positions := []string{"wali_kelas", "bk", "kesiswaan", "kaprog"}
	approverIDs := []id.ID{id.ID(100), id.ID(101), id.ID(102), id.ID(103)}
	steps := make([]pkl.Step, 0, 4)
	for i, pos := range positions {
		steps = append(steps, pkl.Step{
			ID:         id.ID(1000 + i),
			RequestID:  requestID,
			Position:   pos,
			ApproverID: approverIDs[i],
			Status:     pkl.StepPending,
			Sequence:   i + 1,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	return steps
}

func TestService_Create(t *testing.T) {
	start := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC)

	t.Run("success resolves and snapshots approvers", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		users.On("ByID", mock.Anything, id.ID(10)).Return(&user.User{ID: id.ID(10), Class: "PPLG 1", Jurusan: "PPLG"}, nil)
		users.On("FindByPosition", mock.Anything, "wali_kelas", "PPLG 1", "").Return(approver("wali_kelas", id.ID(100)), nil)
		users.On("FindByPosition", mock.Anything, "bk", "", "").Return(approver("bk", id.ID(101)), nil)
		users.On("FindByPosition", mock.Anything, "kesiswaan", "", "").Return(approver("kesiswaan", id.ID(102)), nil)
		users.On("FindByPosition", mock.Anything, "kaprog", "", "PPLG").Return(approver("kaprog", id.ID(103)), nil)
		repo.On("CreateRequest", mock.Anything, mock.MatchedBy(func(r *pkl.PklRequest) bool {
			return r.RequesterID == id.ID(10) && r.Status == pkl.StatusPending && r.CurrentStep == 1 && r.ID != 0
		}), mock.MatchedBy(func(steps []pkl.Step) bool {
			if len(steps) != 4 {
				return false
			}
			for i, s := range steps {
				if s.Position != pkl.ApprovalOrder[i] || s.Sequence != i+1 || s.Status != pkl.StepPending {
					return false
				}
			}
			return steps[0].ApproverID == id.ID(100) && steps[3].ApproverID == id.ID(103)
		})).Return(nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.Create(context.Background(), id.ID(10), "PT Maju", "Jl. Merdeka 1", start, end, "PKL di perusahaan")
		require.NoError(t, err)
		require.Equal(t, pkl.StatusPending, got.Status)
		require.Len(t, got.Steps, 4)
		require.Equal(t, id.ID(100), got.Steps[0].ApproverID)
	})

	t.Run("requester not found", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		users.On("ByID", mock.Anything, id.ID(10)).Return((*user.User)(nil), nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Create(context.Background(), id.ID(10), "PT Maju", "Jl. Merdeka 1", start, end, "PKL")
		require.EqualError(t, err, "requester not found")
	})

	t.Run("no wali kelas assigned", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		users.On("ByID", mock.Anything, id.ID(10)).Return(&user.User{ID: id.ID(10), Class: "PPLG 1", Jurusan: "PPLG"}, nil)
		users.On("FindByPosition", mock.Anything, "wali_kelas", "PPLG 1", "").Return((*user.User)(nil), nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Create(context.Background(), id.ID(10), "PT Maju", "Jl. Merdeka 1", start, end, "PKL")
		require.EqualError(t, err, "no wali_kelas assigned for this request")
	})

	t.Run("no kaprog assigned", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		users.On("ByID", mock.Anything, id.ID(10)).Return(&user.User{ID: id.ID(10), Class: "PPLG 1", Jurusan: "PPLG"}, nil)
		users.On("FindByPosition", mock.Anything, "wali_kelas", "PPLG 1", "").Return(approver("wali_kelas", id.ID(100)), nil)
		users.On("FindByPosition", mock.Anything, "bk", "", "").Return(approver("bk", id.ID(101)), nil)
		users.On("FindByPosition", mock.Anything, "kesiswaan", "", "").Return(approver("kesiswaan", id.ID(102)), nil)
		users.On("FindByPosition", mock.Anything, "kaprog", "", "PPLG").Return((*user.User)(nil), nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Create(context.Background(), id.ID(10), "PT Maju", "Jl. Merdeka 1", start, end, "PKL")
		require.EqualError(t, err, "no kaprog assigned for this request")
	})

	t.Run("end date before start date", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		users.On("ByID", mock.Anything, id.ID(10)).Return(&user.User{ID: id.ID(10), Class: "PPLG 1", Jurusan: "PPLG"}, nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Create(context.Background(), id.ID(10), "PT Maju", "Jl. Merdeka 1", end, start, "PKL")
		require.EqualError(t, err, "end date must be on or after start date")
	})

	t.Run("find by position error", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		users.On("ByID", mock.Anything, id.ID(10)).Return(&user.User{ID: id.ID(10), Class: "PPLG 1", Jurusan: "PPLG"}, nil)
		users.On("FindByPosition", mock.Anything, "wali_kelas", "PPLG 1", "").Return((*user.User)(nil), errors.New("db down"))

		svc := pkl.NewService(repo, users)
		_, err := svc.Create(context.Background(), id.ID(10), "PT Maju", "Jl. Merdeka 1", start, end, "PKL")
		require.EqualError(t, err, "db down")
	})

	t.Run("create request error", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		users.On("ByID", mock.Anything, id.ID(10)).Return(&user.User{ID: id.ID(10), Class: "PPLG 1", Jurusan: "PPLG"}, nil)
		users.On("FindByPosition", mock.Anything, "wali_kelas", "PPLG 1", "").Return(approver("wali_kelas", id.ID(100)), nil)
		users.On("FindByPosition", mock.Anything, "bk", "", "").Return(approver("bk", id.ID(101)), nil)
		users.On("FindByPosition", mock.Anything, "kesiswaan", "", "").Return(approver("kesiswaan", id.ID(102)), nil)
		users.On("FindByPosition", mock.Anything, "kaprog", "", "PPLG").Return(approver("kaprog", id.ID(103)), nil)
		repo.On("CreateRequest", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("insert failed"))

		svc := pkl.NewService(repo, users)
		_, err := svc.Create(context.Background(), id.ID(10), "PT Maju", "Jl. Merdeka 1", start, end, "PKL")
		require.EqualError(t, err, "insert failed")
	})
}

func TestService_List(t *testing.T) {
	t.Run("admin sees all", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		reqs := []pkl.PklRequest{{ID: id.ID(20)}, {ID: id.ID(21)}}
		repo.On("ListAll", mock.Anything).Return(reqs, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(21)).Return(fullSteps(id.ID(21)), nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.List(context.Background(), id.ID(200), "admin")
		require.NoError(t, err)
		require.Len(t, got, 2)
		require.Len(t, got[0].Steps, 4)
	})

	t.Run("guru sees only requests they approve", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		reqs := []pkl.PklRequest{{ID: id.ID(20)}}
		repo.On("ListForApprover", mock.Anything, id.ID(101)).Return(reqs, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.List(context.Background(), id.ID(101), "guru")
		require.NoError(t, err)
		require.Len(t, got, 1)
	})

	t.Run("user sees own requests", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		reqs := []pkl.PklRequest{{ID: id.ID(20)}}
		repo.On("ListByRequester", mock.Anything, id.ID(10)).Return(reqs, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.List(context.Background(), id.ID(10), "user")
		require.NoError(t, err)
		require.Len(t, got, 1)
	})
}

func TestService_Get(t *testing.T) {

	t.Run("requester can view own", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		repo.On("ByID", mock.Anything, id.ID(20)).Return(&pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10)}, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.Get(context.Background(), id.ID(10), "user", id.ID(20))
		require.NoError(t, err)
		require.Equal(t, id.ID(20), got.ID)
	})

	t.Run("other user forbidden", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		repo.On("ByID", mock.Anything, id.ID(20)).Return(&pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10)}, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Get(context.Background(), id.ID(11), "user", id.ID(20))
		require.EqualError(t, err, "forbidden: not the requester")
	})

	t.Run("approver guru can view", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		repo.On("ByID", mock.Anything, id.ID(20)).Return(&pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10)}, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.Get(context.Background(), id.ID(100), "guru", id.ID(20))
		require.NoError(t, err)
		require.Equal(t, id.ID(20), got.ID)
	})

	t.Run("unrelated guru forbidden", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		repo.On("ByID", mock.Anything, id.ID(20)).Return(&pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10)}, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Get(context.Background(), id.ID(104), "guru", id.ID(20))
		require.EqualError(t, err, "forbidden: not an approver on this request")
	})

	t.Run("admin can view any", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		repo.On("ByID", mock.Anything, id.ID(20)).Return(&pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10)}, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.Get(context.Background(), id.ID(200), "admin", id.ID(20))
		require.NoError(t, err)
		require.Equal(t, id.ID(20), got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		repo.On("ByID", mock.Anything, id.ID(99)).Return((*pkl.PklRequest)(nil), nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Get(context.Background(), id.ID(200), "admin", id.ID(99))
		require.EqualError(t, err, "pkl request not found")
	})
}

func TestService_Decide(t *testing.T) {

	t.Run("approve advances to next step", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusPending, CurrentStep: 1}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)
		repo.On("Decide", mock.Anything, mock.MatchedBy(func(r *pkl.PklRequest) bool {
			return r.Status == pkl.StatusPending && r.CurrentStep == 2
		}), mock.MatchedBy(func(s *pkl.Step) bool {
			return s.Position == "wali_kelas" && s.Status == pkl.StepApproved
		}), pkl.StatusPending, pkl.StepPending).Return(nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.Decide(context.Background(), id.ID(100), id.ID(20), pkl.DecisionApprove, "ok")
		require.NoError(t, err)
		require.Equal(t, pkl.StatusPending, got.Status)
		require.Equal(t, 2, got.CurrentStep)
	})

	t.Run("approve last step accepts request", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusPending, CurrentStep: 4}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)
		repo.On("Decide", mock.Anything, mock.MatchedBy(func(r *pkl.PklRequest) bool {
			return r.Status == pkl.StatusAccepted
		}), mock.MatchedBy(func(s *pkl.Step) bool {
			return s.Position == "kaprog" && s.Status == pkl.StepApproved
		}), pkl.StatusPending, pkl.StepPending).Return(nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.Decide(context.Background(), id.ID(103), id.ID(20), pkl.DecisionApprove, "")
		require.NoError(t, err)
		require.Equal(t, pkl.StatusAccepted, got.Status)
	})

	t.Run("reject closes request", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusPending, CurrentStep: 1}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)
		repo.On("Decide", mock.Anything, mock.MatchedBy(func(r *pkl.PklRequest) bool {
			return r.Status == pkl.StatusRejected
		}), mock.MatchedBy(func(s *pkl.Step) bool {
			return s.Status == pkl.StepRejected
		}), pkl.StatusPending, pkl.StepPending).Return(nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.Decide(context.Background(), id.ID(100), id.ID(20), pkl.DecisionReject, "tidak sesuai")
		require.NoError(t, err)
		require.Equal(t, pkl.StatusRejected, got.Status)
	})

	t.Run("needs further action freezes request", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusPending, CurrentStep: 2}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)
		repo.On("Decide", mock.Anything, mock.MatchedBy(func(r *pkl.PklRequest) bool {
			return r.Status == pkl.StatusNeedsAction
		}), mock.MatchedBy(func(s *pkl.Step) bool {
			return s.Position == "bk" && s.Status == pkl.StepNeedsAction
		}), pkl.StatusPending, pkl.StepPending).Return(nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.Decide(context.Background(), id.ID(101), id.ID(20), pkl.DecisionNeedsAction, "datang ke ruang BK")
		require.NoError(t, err)
		require.Equal(t, pkl.StatusNeedsAction, got.Status)
		require.Equal(t, 2, got.CurrentStep)
	})

	t.Run("resolve needs further action with approve", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		frozen := fullSteps(id.ID(20))
		frozen[1].Status = pkl.StepNeedsAction
		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusNeedsAction, CurrentStep: 2}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(frozen, nil)
		repo.On("Decide", mock.Anything, mock.MatchedBy(func(r *pkl.PklRequest) bool {
			return r.Status == pkl.StatusPending && r.CurrentStep == 3
		}), mock.MatchedBy(func(s *pkl.Step) bool {
			return s.Position == "bk" && s.Status == pkl.StepApproved
		}), pkl.StatusNeedsAction, pkl.StepNeedsAction).Return(nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.Decide(context.Background(), id.ID(101), id.ID(20), pkl.DecisionApprove, "sudah bertemu")
		require.NoError(t, err)
		require.Equal(t, pkl.StatusPending, got.Status)
		require.Equal(t, 3, got.CurrentStep)
	})

	t.Run("resolve needs further action with reject", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		frozen := fullSteps(id.ID(20))
		frozen[1].Status = pkl.StepNeedsAction
		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusNeedsAction, CurrentStep: 2}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(frozen, nil)
		repo.On("Decide", mock.Anything, mock.MatchedBy(func(r *pkl.PklRequest) bool {
			return r.Status == pkl.StatusRejected
		}), mock.MatchedBy(func(s *pkl.Step) bool {
			return s.Status == pkl.StepRejected
		}), pkl.StatusNeedsAction, pkl.StepNeedsAction).Return(nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.Decide(context.Background(), id.ID(101), id.ID(20), pkl.DecisionReject, "tidak lanjut")
		require.NoError(t, err)
		require.Equal(t, pkl.StatusRejected, got.Status)
	})

	t.Run("wrong guru at pending step forbidden", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusPending, CurrentStep: 2}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Decide(context.Background(), id.ID(102), id.ID(20), pkl.DecisionApprove, "")
		require.EqualError(t, err, "forbidden: not your step to decide")
	})

	t.Run("not assigned guru cannot resolve needs further action", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		frozen := fullSteps(id.ID(20))
		frozen[1].Status = pkl.StepNeedsAction
		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusNeedsAction, CurrentStep: 2}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(frozen, nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Decide(context.Background(), id.ID(103), id.ID(20), pkl.DecisionApprove, "")
		require.EqualError(t, err, "forbidden: not your step to decide")
	})

	t.Run("request not awaiting decision", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusAccepted, CurrentStep: 4}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Decide(context.Background(), id.ID(103), id.ID(20), pkl.DecisionApprove, "")
		require.EqualError(t, err, "request is not awaiting a decision")
	})

	t.Run("invalid decision", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusPending, CurrentStep: 1}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Decide(context.Background(), id.ID(100), id.ID(20), "maybe", "")
		require.EqualError(t, err, "invalid decision: must be approve, reject or needs_further_action")
	})

	t.Run("needs further action on already frozen request", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		frozen := fullSteps(id.ID(20))
		frozen[1].Status = pkl.StepNeedsAction
		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusNeedsAction, CurrentStep: 2}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(frozen, nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Decide(context.Background(), id.ID(101), id.ID(20), pkl.DecisionNeedsAction, "")
		require.EqualError(t, err, "request already needs further action")
	})

	t.Run("request not found", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		repo.On("ByID", mock.Anything, id.ID(99)).Return((*pkl.PklRequest)(nil), nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Decide(context.Background(), id.ID(100), id.ID(99), pkl.DecisionApprove, "")
		require.EqualError(t, err, "pkl request not found")
	})

	t.Run("repository decide error", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusPending, CurrentStep: 1}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("StepsByRequest", mock.Anything, id.ID(20)).Return(fullSteps(id.ID(20)), nil)
		repo.On("Decide", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("conflict"))

		svc := pkl.NewService(repo, users)
		_, err := svc.Decide(context.Background(), id.ID(100), id.ID(20), pkl.DecisionApprove, "")
		require.EqualError(t, err, "conflict")
	})
}

func TestService_Cancel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusPending}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("Cancel", mock.Anything, mock.MatchedBy(func(r *pkl.PklRequest) bool {
			return r.Status == pkl.StatusCancelled && r.CancelReason == "salah perusahaan"
		})).Return(nil)

		svc := pkl.NewService(repo, users)
		got, err := svc.Cancel(context.Background(), id.ID(10), id.ID(20), "salah perusahaan")
		require.NoError(t, err)
		require.Equal(t, pkl.StatusCancelled, got.Status)
	})

	t.Run("reason required", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		svc := pkl.NewService(repo, users)
		_, err := svc.Cancel(context.Background(), id.ID(10), id.ID(20), "")
		require.EqualError(t, err, "cancellation reason is required")
	})

	t.Run("not found", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		repo.On("ByID", mock.Anything, id.ID(99)).Return((*pkl.PklRequest)(nil), nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Cancel(context.Background(), id.ID(10), id.ID(99), "alasan")
		require.EqualError(t, err, "pkl request not found")
	})

	t.Run("not requester forbidden", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusPending}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Cancel(context.Background(), id.ID(11), id.ID(20), "alasan")
		require.EqualError(t, err, "forbidden: not the requester")
	})

	t.Run("accepted request cannot be cancelled", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusAccepted}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)

		svc := pkl.NewService(repo, users)
		_, err := svc.Cancel(context.Background(), id.ID(10), id.ID(20), "alasan")
		require.EqualError(t, err, "only pending or needs_further_action requests can be cancelled")
	})

	t.Run("repository error", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		users := mocks.NewUsers(t)

		req := &pkl.PklRequest{ID: id.ID(20), RequesterID: id.ID(10), Status: pkl.StatusPending}
		repo.On("ByID", mock.Anything, id.ID(20)).Return(req, nil)
		repo.On("Cancel", mock.Anything, mock.Anything).Return(errors.New("cancel failed"))

		svc := pkl.NewService(repo, users)
		_, err := svc.Cancel(context.Background(), id.ID(10), id.ID(20), "alasan")
		require.EqualError(t, err, "cancel failed")
	})
}
