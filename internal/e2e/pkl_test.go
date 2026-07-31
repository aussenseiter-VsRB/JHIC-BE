//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/pkl"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"github.com/stretchr/testify/require"
)

type guruAccount struct {
	id    id.ID
	token string
}

func createUserAsAdmin(t *testing.T, e *env, adminToken, email, role, class, jurusan, position string) id.ID {
	t.Helper()
	body := map[string]string{
		"email":    email,
		"password": "secret-password",
		"name":     email,
		"role":     role,
		"class":    class,
		"jurusan":  jurusan,
		"position": position,
	}
	resp := doJSON(t, http.MethodPost, e.server.URL+"/api/v1/users", adminToken, body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var out auth.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	resp.Body.Close()
	return out.ID
}

func login(t *testing.T, e *env, email string) string {
	t.Helper()
	resp := doJSON(t, http.MethodPost, e.server.URL+"/api/v1/auth/login", "", map[string]string{
		"email":    email,
		"password": "secret-password",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	resp.Body.Close()
	require.NotEmpty(t, out.Token)
	return out.Token
}

func seedPKLApprovers(t *testing.T, e *env, adminToken string) map[string]guruAccount {
	t.Helper()
	gurus := map[string]guruAccount{}
	positions := map[string]struct {
		class   string
		jurusan string
	}{
		"wali_kelas": {class: "PPLG 1"},
		"bk":         {},
		"kesiswaan":  {},
		"kaprog":     {jurusan: "PPLG"},
	}
	for position, meta := range positions {
		uid := createUserAsAdmin(t, e, adminToken, position+"@example.com", "guru", meta.class, meta.jurusan, position)
		gurus[position] = guruAccount{id: uid, token: login(t, e, position+"@example.com")}
	}
	return gurus
}

func createPKLRequest(t *testing.T, e *env, studentToken string) pkl.PklRequest {
	t.Helper()
	resp := doJSON(t, http.MethodPost, e.server.URL+"/api/v1/approval/pkl", studentToken, map[string]string{
		"company":     "PT Maju",
		"location":    "Jl. Merdeka 1",
		"start_date":  "2026-08-03",
		"end_date":    "2026-08-28",
		"description": "PKL di bagian IT",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var req pkl.PklRequest
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&req))
	resp.Body.Close()
	return req
}

func decide(t *testing.T, e *env, guruToken string, requestID id.ID, decision, note string) *http.Response {
	t.Helper()
	body := map[string]string{"decision": decision}
	if note != "" {
		body["note"] = note
	}
	return doJSON(t, http.MethodPost, e.server.URL+"/api/v1/approval/pkl/"+requestID.String()+"/decide", guruToken, body)
}

func TestE2E_PKLFullApprovalFlow(t *testing.T) {
	e := startE2E(t)
	ctx := t.Context()

	adminID, adminToken := register(t, e.server.URL, "admin@example.com")
	promoteToAdmin(t, e, adminID)

	gurus := seedPKLApprovers(t, e, adminToken)
	createUserAsAdmin(t, e, adminToken, "student@example.com", "user", "PPLG 1", "PPLG", "")
	studentToken := login(t, e, "student@example.com")

	req := createPKLRequest(t, e, studentToken)
	require.Equal(t, pkl.StatusPending, req.Status)
	require.Equal(t, 1, req.CurrentStep)
	require.Len(t, req.Steps, 4)
	for i, s := range req.Steps {
		require.Equal(t, pkl.ApprovalOrder[i], s.Position)
		require.Equal(t, req.ID, s.RequestID)
		require.Equal(t, pkl.StepPending, s.Status)
	}

	var count int
	require.NoError(t, e.pool.QueryRow(ctx, `SELECT COUNT(*) FROM pkl_approval_steps WHERE pkl_request_id = $1`, req.ID).Scan(&count))
	require.Equal(t, 4, count)

	listResp := doJSON(t, http.MethodGet, e.server.URL+"/api/v1/approval/pkl", studentToken, nil)
	require.Equal(t, http.StatusOK, listResp.StatusCode)
	var studentList []pkl.PklRequest
	require.NoError(t, json.NewDecoder(listResp.Body).Decode(&studentList))
	listResp.Body.Close()
	require.Len(t, studentList, 1)

	for _, guru := range gurus {
		resp := doJSON(t, http.MethodGet, e.server.URL+"/api/v1/approval/pkl", guru.token, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var list []pkl.PklRequest
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
		resp.Body.Close()
		require.Len(t, list, 1)
	}

	adminListResp := doJSON(t, http.MethodGet, e.server.URL+"/api/v1/approval/pkl", adminToken, nil)
	require.Equal(t, http.StatusOK, adminListResp.StatusCode)
	var adminList []pkl.PklRequest
	require.NoError(t, json.NewDecoder(adminListResp.Body).Decode(&adminList))
	adminListResp.Body.Close()
	require.Len(t, adminList, 1)

	outOfTurn := decide(t, e, gurus["bk"].token, req.ID, "approve", "")
	require.Equal(t, http.StatusForbidden, outOfTurn.StatusCode)
	outOfTurn.Body.Close()

	for _, position := range pkl.ApprovalOrder[:3] {
		resp := decide(t, e, gurus[position].token, req.ID, "approve", "setuju")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var got pkl.PklRequest
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		resp.Body.Close()
		require.Equal(t, req.ID, got.ID)
		require.Equal(t, pkl.StatusPending, got.Status)
		require.Equal(t, got.CurrentStep, positionIndex(position)+2)
	}

	resp := decide(t, e, gurus["kaprog"].token, req.ID, "approve", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var accepted pkl.PklRequest
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&accepted))
	resp.Body.Close()
	require.Equal(t, pkl.StatusAccepted, accepted.Status)
	require.Equal(t, 4, accepted.CurrentStep)

	again := decide(t, e, gurus["kaprog"].token, req.ID, "approve", "")
	require.Equal(t, http.StatusConflict, again.StatusCode)
	again.Body.Close()

	cancelAccepted := doJSON(t, http.MethodDelete, e.server.URL+"/api/v1/approval/pkl/"+req.ID.String(), studentToken, map[string]string{"reason": "tidak jadi"})
	require.Equal(t, http.StatusConflict, cancelAccepted.StatusCode)
	cancelAccepted.Body.Close()

	var status string
	require.NoError(t, e.pool.QueryRow(ctx, `SELECT status FROM pkl_requests WHERE id = $1`, req.ID).Scan(&status))
	require.Equal(t, pkl.StatusAccepted, status)
}

func positionIndex(position string) int {
	for i, p := range pkl.ApprovalOrder {
		if p == position {
			return i
		}
	}
	return -1
}

func TestE2E_PKLNeedsFurtherAction(t *testing.T) {
	e := startE2E(t)

	adminID, adminToken := register(t, e.server.URL, "admin@example.com")
	promoteToAdmin(t, e, adminID)

	gurus := seedPKLApprovers(t, e, adminToken)
	createUserAsAdmin(t, e, adminToken, "student@example.com", "user", "PPLG 1", "PPLG", "")
	studentToken := login(t, e, "student@example.com")

	req := createPKLRequest(t, e, studentToken)

	resp := decide(t, e, gurus["bk"].token, req.ID, "needs_further_action", "")
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()

	resp = decide(t, e, gurus["wali_kelas"].token, req.ID, "needs_further_action", "tolong temui saya")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var needsAction pkl.PklRequest
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&needsAction))
	resp.Body.Close()
	require.Equal(t, pkl.StatusNeedsAction, needsAction.Status)
	require.Equal(t, 1, needsAction.CurrentStep)

	dup := decide(t, e, gurus["wali_kelas"].token, req.ID, "needs_further_action", "")
	require.Equal(t, http.StatusConflict, dup.StatusCode)
	dup.Body.Close()

	nextStep := decide(t, e, gurus["bk"].token, req.ID, "approve", "")
	require.Equal(t, http.StatusForbidden, nextStep.StatusCode)
	nextStep.Body.Close()

	resp = decide(t, e, gurus["wali_kelas"].token, req.ID, "approve", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var resumed pkl.PklRequest
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&resumed))
	resp.Body.Close()
	require.Equal(t, pkl.StatusPending, resumed.Status)
	require.Equal(t, 2, resumed.CurrentStep)

	noReason := doJSON(t, http.MethodDelete, e.server.URL+"/api/v1/approval/pkl/"+req.ID.String(), studentToken, map[string]string{})
	require.Equal(t, http.StatusBadRequest, noReason.StatusCode)
	noReason.Body.Close()

	cancelled := doJSON(t, http.MethodDelete, e.server.URL+"/api/v1/approval/pkl/"+req.ID.String(), studentToken, map[string]string{"reason": "mundur"})
	require.Equal(t, http.StatusOK, cancelled.StatusCode)
	var cancelledReq pkl.PklRequest
	require.NoError(t, json.NewDecoder(cancelled.Body).Decode(&cancelledReq))
	cancelled.Body.Close()
	require.Equal(t, pkl.StatusCancelled, cancelledReq.Status)
	require.Equal(t, "mundur", cancelledReq.CancelReason)
}

func TestE2E_PKLReject(t *testing.T) {
	e := startE2E(t)
	ctx := t.Context()

	adminID, adminToken := register(t, e.server.URL, "admin@example.com")
	promoteToAdmin(t, e, adminID)

	gurus := seedPKLApprovers(t, e, adminToken)
	createUserAsAdmin(t, e, adminToken, "student@example.com", "user", "PPLG 1", "PPLG", "")
	studentToken := login(t, e, "student@example.com")

	req := createPKLRequest(t, e, studentToken)

	resp := decide(t, e, gurus["wali_kelas"].token, req.ID, "approve", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = decide(t, e, gurus["bk"].token, req.ID, "reject", "dukumen kurang")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var rejected pkl.PklRequest
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rejected))
	resp.Body.Close()
	require.Equal(t, pkl.StatusRejected, rejected.Status)

	frozen := decide(t, e, gurus["kesiswaan"].token, req.ID, "approve", "")
	require.Equal(t, http.StatusConflict, frozen.StatusCode)
	frozen.Body.Close()

	cancelRejected := doJSON(t, http.MethodDelete, e.server.URL+"/api/v1/approval/pkl/"+req.ID.String(), studentToken, map[string]string{"reason": "tidak jadi"})
	require.Equal(t, http.StatusConflict, cancelRejected.StatusCode)
	cancelRejected.Body.Close()

	require.NoError(t, e.pool.QueryRow(ctx, `SELECT status FROM pkl_requests WHERE id = $1`, req.ID).Scan(&rejected.Status))
	require.Equal(t, pkl.StatusRejected, rejected.Status)
}

func TestE2E_PKLRoleAndVisibility(t *testing.T) {
	e := startE2E(t)

	adminID, adminToken := register(t, e.server.URL, "admin@example.com")
	promoteToAdmin(t, e, adminID)

	gurus := seedPKLApprovers(t, e, adminToken)
	createUserAsAdmin(t, e, adminToken, "student@example.com", "user", "PPLG 1", "PPLG", "")
	studentToken := login(t, e, "student@example.com")

	createUserAsAdmin(t, e, adminToken, "other@example.com", "user", "HTL 1", "HTL", "")
	otherStudentToken := login(t, e, "other@example.com")

	foreignKaprogID := createUserAsAdmin(t, e, adminToken, "kaprog-htl@example.com", "guru", "", "HTL", "kaprog")
	foreignKaprogToken := login(t, e, "kaprog-htl@example.com")
	require.NotEqual(t, gurus["kaprog"].id, foreignKaprogID)

	req := createPKLRequest(t, e, studentToken)

	otherSeesOwn := doJSON(t, http.MethodGet, e.server.URL+"/api/v1/approval/pkl", otherStudentToken, nil)
	require.Equal(t, http.StatusOK, otherSeesOwn.StatusCode)
	var otherList []pkl.PklRequest
	require.NoError(t, json.NewDecoder(otherSeesOwn.Body).Decode(&otherList))
	otherSeesOwn.Body.Close()
	require.Len(t, otherList, 0)

	forbiddenGet := doJSON(t, http.MethodGet, e.server.URL+"/api/v1/approval/pkl/"+req.ID.String(), otherStudentToken, nil)
	require.Equal(t, http.StatusForbidden, forbiddenGet.StatusCode)
	forbiddenGet.Body.Close()

	foreignGet := doJSON(t, http.MethodGet, e.server.URL+"/api/v1/approval/pkl/"+req.ID.String(), foreignKaprogToken, nil)
	require.Equal(t, http.StatusForbidden, foreignGet.StatusCode)
	foreignGet.Body.Close()

	adminGet := doJSON(t, http.MethodGet, e.server.URL+"/api/v1/approval/pkl/"+req.ID.String(), adminToken, nil)
	require.Equal(t, http.StatusOK, adminGet.StatusCode)
	adminGet.Body.Close()

	studentDecides := decide(t, e, studentToken, req.ID, "approve", "")
	require.Equal(t, http.StatusForbidden, studentDecides.StatusCode)
	studentDecides.Body.Close()

	adminDecides := decide(t, e, adminToken, req.ID, "approve", "")
	require.Equal(t, http.StatusForbidden, adminDecides.StatusCode)
	adminDecides.Body.Close()

	noAuth := doJSON(t, http.MethodGet, e.server.URL+"/api/v1/approval/pkl", "", nil)
	require.Equal(t, http.StatusUnauthorized, noAuth.StatusCode)
	noAuth.Body.Close()

	missing := doJSON(t, http.MethodGet, e.server.URL+"/api/v1/approval/pkl/999999999999999999", studentToken, nil)
	require.Equal(t, http.StatusNotFound, missing.StatusCode)
	missing.Body.Close()
}
