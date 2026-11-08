package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Polikarpov00/test_task_wallet/internal/domain"
	"github.com/google/uuid"
)

type fakeRepo struct {
	called bool
}

func (f *fakeRepo) ApplyOperation(_ context.Context, _ uuid.UUID, _ domain.OperationType, _ int64) (int64, error) {
	f.called = true
	return 100, nil
}

func (f *fakeRepo) GetBalance(_ context.Context, _ uuid.UUID) (int64, error) {
	return 100, nil
}

// Валидация входных данных перед обращением к репозиторию.
// Пустой id, неизвестный тип операции, amount <= 0 — ошибка, репозиторий не вызывается.
// Валидный запрос — без ошибки, репозиторий вызывается.
func TestApplyOperation_Validation(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		op      domain.OperationType
		amount  int64
		wantErr error
	}{
		{"valid deposit", validID, domain.Deposit, 100, nil},
		{"nil wallet id", uuid.Nil, domain.Deposit, 100, domain.ErrInvalidWalletID},
		{"unknown operation", validID, "deposit", 100, domain.ErrInvalidOperation}, // регистр важен
		{"zero amount", validID, domain.Withdraw, 0, domain.ErrInvalidAmount},
		{"negative amount", validID, domain.Deposit, -50, domain.ErrInvalidAmount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{}
			svc := NewWalletService(repo)

			_, err := svc.ApplyOperation(context.Background(), tt.id, tt.op, tt.amount)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil && repo.called {
				t.Error("repository must not be called on invalid input")
			}
			if tt.wantErr == nil && !repo.called {
				t.Error("repository was not called on valid input")
			}
		})
	}
}
