package cvreview_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/cvreview"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/cvreview/content"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const validCvJSON = `{
	"audit_summary": {
		"score": 80,
		"tier_label": "Kandidat Kuat",
		"grade_label": "B+",
		"summary_text": "Ringkasan.",
		"key_strengths": ["pengalaman"],
		"key_improvements": ["kata kunci"]
	},
	"metrics": {"format_score": 85, "ats_status": "good"},
	"grammar_issues": [],
	"recommendations": [],
	"strengths_detail": []
}`

func TestService_CvReview(t *testing.T) {
	t.Run("success forwards and normalizes", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		client.On("CvReview", mock.Anything, "CV saya", 5, 1).Return(validCvJSON, nil)

		svc := cvreview.NewService(client)
		got, err := svc.CvReview(context.Background(), cvreview.CvReviewRequest{
			CvText:    "   CV saya  ",
			WordCount: 5,
			PageCount: 1,
		})
		require.NoError(t, err)
		require.Equal(t, 80, got.AuditSummary.Score)
		require.Equal(t, "good", got.Metrics.ATSStatus)
	})

	t.Run("empty cv_text rejected before upstream call", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := cvreview.NewService(client)
		_, err := svc.CvReview(context.Background(), cvreview.CvReviewRequest{CvText: "   "})
		require.ErrorIs(t, err, cvreview.ErrCvTextRequired)
	})

	t.Run("overlong cv_text rejected", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := cvreview.NewService(client)
		_, err := svc.CvReview(context.Background(), cvreview.CvReviewRequest{
			CvText: strings.Repeat("a", content.CvTextMaxLen+1),
		})
		require.ErrorIs(t, err, cvreview.ErrCvTextTooLong)
	})

	t.Run("negative counts rejected", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := cvreview.NewService(client)
		_, err := svc.CvReview(context.Background(), cvreview.CvReviewRequest{
			CvText:    "cv",
			WordCount: -1,
		})
		require.ErrorIs(t, err, cvreview.ErrInvalidCounts)
	})

	t.Run("invalid model output rejected", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		client.On("CvReview", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("bukan json", nil)

		svc := cvreview.NewService(client)
		_, err := svc.CvReview(context.Background(), cvreview.CvReviewRequest{CvText: "cv"})
		require.ErrorIs(t, err, cvreview.ErrCvOutputInvalid)
	})
}