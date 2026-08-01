package ai_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/ai"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/ai/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Chat(t *testing.T) {
	t.Run("success forwards message and session", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		client.On("Chat", mock.Anything, "halo", "session-1").Return(&ai.ChatResponse{Output: "hai"}, nil)

		svc := ai.NewService(client)
		got, err := svc.Chat(context.Background(), "halo", "session-1")
		require.NoError(t, err)
		require.Equal(t, "hai", got.Output)
	})

	t.Run("trims input and forwards empty session", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		client.On("Chat", mock.Anything, "halo", mock.MatchedBy(func(sid string) bool {
			return len(sid) == 32
		})).Return(&ai.ChatResponse{Output: "hai"}, nil)

		svc := ai.NewService(client)
		got, err := svc.Chat(context.Background(), "  halo  ", "")
		require.NoError(t, err)
		require.Equal(t, "hai", got.Output)
	})

	t.Run("empty message rejected", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := ai.NewService(client)
		_, err := svc.Chat(context.Background(), "   ", "session-1")
		require.ErrorIs(t, err, ai.ErrChatMessageRequired)
	})

	t.Run("overlong message rejected", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := ai.NewService(client)
		_, err := svc.Chat(context.Background(), strings.Repeat("a", ai.ChatMessageMaxLen+1), "session-1")
		require.ErrorIs(t, err, ai.ErrChatMessageTooLong)
	})

	t.Run("exactly max length accepted", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		client.On("Chat", mock.Anything, mock.MatchedBy(func(s string) bool {
			return len(s) == ai.ChatMessageMaxLen
		}), "session-1").Return(&ai.ChatResponse{Output: "hai"}, nil)

		svc := ai.NewService(client)
		_, err := svc.Chat(context.Background(), strings.Repeat("a", ai.ChatMessageMaxLen), "session-1")
		require.NoError(t, err)
	})

	t.Run("propagates upstream error", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		client.On("Chat", mock.Anything, "halo", "session-1").Return(nil, ai.ErrN8NUnavailable)

		svc := ai.NewService(client)
		_, err := svc.Chat(context.Background(), "halo", "session-1")
		require.ErrorIs(t, err, ai.ErrN8NUnavailable)
	})
}

func TestService_NexxaMatch(t *testing.T) {
	valid := []string{"a", "b", "c", "d", "e", "f", "g", "h"}

	t.Run("success forwards normalized answers", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		client.On("NexxaMatch", mock.Anything, mock.MatchedBy(func(a []string) bool {
			if len(a) != ai.NexxaAnswerCount {
				return false
			}
			for i, v := range a {
				if v != valid[i] || strings.TrimSpace(v) != v {
					return false
				}
			}
			return true
		})).Return(&ai.NexxaResponse{NamaJurusan: "PPLG", Alasan: "cocok"}, nil)

		svc := ai.NewService(client)
		got, err := svc.NexxaMatch(context.Background(), []string{" a ", "b", "c", "d", "e", "f", "g", "h"})
		require.NoError(t, err)
		require.Equal(t, "PPLG", got.NamaJurusan)
		require.Equal(t, "cocok", got.Alasan)
	})

	t.Run("wrong answer count rejected", func(t *testing.T) {
		for _, answers := range [][]string{
			{"a", "b", "c", "d", "e", "f", "g"},
			append(valid, "i"),
		} {
			client := mocks.NewN8NClient(t)
			svc := ai.NewService(client)
			_, err := svc.NexxaMatch(context.Background(), answers)
			require.ErrorIs(t, err, ai.ErrAnswersRequired)
		}
	})

	t.Run("empty answer rejected", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := ai.NewService(client)
		_, err := svc.NexxaMatch(context.Background(), []string{"a", "b", "  ", "d", "e", "f", "g", "h"})
		require.ErrorIs(t, err, ai.ErrAnswersRequired)
	})

	t.Run("overlong answer rejected", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := ai.NewService(client)
		answers := append([]string(nil), valid...)
		answers[0] = strings.Repeat("a", ai.NexxaAnswerMaxLen+1)
		_, err := svc.NexxaMatch(context.Background(), answers)
		require.ErrorIs(t, err, ai.ErrAnswerTooLong)
	})
}
