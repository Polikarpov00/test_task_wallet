package service

import (
	"context"

	"github.com/Polikarpov00/test_task_wallet/internal/domain"
	"github.com/google/uuid"
)

type WalletRepository interface {
	ApplyOperation(ctx context.Context, id uuid.UUID, op domain.OperationType, amount int64) (int64, error)
	GetBalance(ctx context.Context, id uuid.UUID) (int64, error)
}

type WalletService struct {
	repo WalletRepository
}

func NewWalletService(repo WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) ApplyOperation(ctx context.Context, id uuid.UUID, op domain.OperationType, amount int64) (int64, error) {
	if id == uuid.Nil {
		return 0, domain.ErrInvalidWalletID
	}
	if !op.Valid() {
		return 0, domain.ErrInvalidOperation
	}
	if amount <= 0 {
		return 0, domain.ErrInvalidAmount
	}
	return s.repo.ApplyOperation(ctx, id, op, amount)
}

func (s *WalletService) GetBalance(ctx context.Context, id uuid.UUID) (int64, error) {
	return s.repo.GetBalance(ctx, id)
}
