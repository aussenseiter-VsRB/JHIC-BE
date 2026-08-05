package spmb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/spmb"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/spmb/mocks"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := spmb.NewService(mockRepo)

	input := validInput()
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *spmb.SpmbRegistration) bool {
		return r.Status == spmb.StatusProses && r.ID != 0
	})).Return(nil)

	got, err := svc.Create(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, spmb.StatusProses, got.Status)
	require.NotZero(t, got.ID)
	require.False(t, got.CreatedAt.IsZero())
	mockRepo.AssertExpectations(t)
}

func TestService_CreateValidation(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*spmb.SpmbRegistration)
		err  string
	}{
		{name: "nama required", mut: func(r *spmb.SpmbRegistration) { r.Nama = "" }, err: "nama is required"},
		{name: "nik length", mut: func(r *spmb.SpmbRegistration) { r.Nik = "123" }, err: "nik must be 16 digits"},
		{name: "jenis kelamin required", mut: func(r *spmb.SpmbRegistration) { r.JenisKelamin = "" }, err: "jenis_kelamin is required"},
		{name: "alamat required", mut: func(r *spmb.SpmbRegistration) { r.Alamat = "" }, err: "alamat is required"},
		{name: "jurusan required", mut: func(r *spmb.SpmbRegistration) { r.Jurusan = "" }, err: "jurusan is required"},
		{name: "jurusan invalid", mut: func(r *spmb.SpmbRegistration) { r.Jurusan = "Fisika" }, err: "invalid jurusan"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			svc := spmb.NewService(mockRepo)
			input := validInput()
			tc.mut(input)
			_, err := svc.Create(context.Background(), input)
			require.EqualError(t, err, tc.err)
			mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		})
	}
}

func TestService_CreateRepoError(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := spmb.NewService(mockRepo)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(errRepo)

	_, err := svc.Create(context.Background(), validInput())
	require.ErrorIs(t, err, errRepo)
}

func TestService_Get(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := spmb.NewService(mockRepo)

	reg := validInput()
	mockRepo.On("ByID", mock.Anything, reg.ID).Return(reg, nil)
	got, err := svc.Get(context.Background(), reg.ID)
	require.NoError(t, err)
	require.Equal(t, reg.ID, got.ID)

	mockRepo.On("ByID", mock.Anything, mock.Anything).Return((*spmb.SpmbRegistration)(nil), nil)
	_, err = svc.Get(context.Background(), id.New())
	require.EqualError(t, err, "registration not found")
}

func TestService_SetStatusApprove(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := spmb.NewService(mockRepo)

	reg := validInput()
	reg.Status = spmb.StatusProses
	mockRepo.On("ByID", mock.Anything, reg.ID).Return(reg, nil)
	mockRepo.On("UpdateStatus", mock.Anything, mock.MatchedBy(func(r *spmb.SpmbRegistration) bool {
		return r.Status == spmb.StatusApprove
	}), spmb.StatusProses).Return(nil)

	got, err := svc.SetStatus(context.Background(), reg.ID, spmb.StatusApprove, "")
	require.NoError(t, err)
	require.Equal(t, spmb.StatusApprove, got.Status)
	mockRepo.AssertExpectations(t)
}

func TestService_SetStatusApproveWrongState(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := spmb.NewService(mockRepo)

	reg := validInput()
	reg.Status = spmb.StatusCancel
	mockRepo.On("ByID", mock.Anything, reg.ID).Return(reg, nil)

	_, err := svc.SetStatus(context.Background(), reg.ID, spmb.StatusApprove, "")
	require.EqualError(t, err, "only proses registrations can be approved")
}

func TestService_SetStatusCancel(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := spmb.NewService(mockRepo)

	reg := validInput()
	reg.Status = spmb.StatusProses
	mockRepo.On("ByID", mock.Anything, reg.ID).Return(reg, nil)
	mockRepo.On("UpdateStatus", mock.Anything, mock.MatchedBy(func(r *spmb.SpmbRegistration) bool {
		return r.Status == spmb.StatusCancel && r.CancelReason == "tidak melengkapi berkas"
	}), spmb.StatusProses).Return(nil)

	got, err := svc.SetStatus(context.Background(), reg.ID, spmb.StatusCancel, "tidak melengkapi berkas")
	require.NoError(t, err)
	require.Equal(t, spmb.StatusCancel, got.Status)
	require.Equal(t, "tidak melengkapi berkas", got.CancelReason)
	mockRepo.AssertExpectations(t)
}

func TestService_SetStatusCancelRequiresReason(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := spmb.NewService(mockRepo)

	reg := validInput()
	reg.Status = spmb.StatusProses
	mockRepo.On("ByID", mock.Anything, reg.ID).Return(reg, nil)

	_, err := svc.SetStatus(context.Background(), reg.ID, spmb.StatusCancel, "")
	require.EqualError(t, err, "cancellation reason is required")
}

func TestService_SetStatusInvalid(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := spmb.NewService(mockRepo)

	reg := validInput()
	mockRepo.On("ByID", mock.Anything, reg.ID).Return(reg, nil)

	_, err := svc.SetStatus(context.Background(), reg.ID, "bogus", "")
	require.EqualError(t, err, "invalid status: must be approve or cancel")
}

func TestService_SetStatusNotFound(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := spmb.NewService(mockRepo)
	mockRepo.On("ByID", mock.Anything, mock.Anything).Return((*spmb.SpmbRegistration)(nil), nil)

	_, err := svc.SetStatus(context.Background(), id.New(), spmb.StatusApprove, "")
	require.EqualError(t, err, "registration not found")
}

var errRepo = errors.New("repo failure")

func validInput() *spmb.SpmbRegistration {
	return &spmb.SpmbRegistration{
		ID:           id.New(),
		Nama:         "Budi Santoso",
		Nik:          "3204123456789012",
		JenisKelamin: "Laki-laki",
		Alamat:       "Jl. Raya Soreang No. 1",
		Jurusan:      "PPLG",
	}
}
