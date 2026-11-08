package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Polikarpov00/test_task_wallet/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepo struct {
	pool *pgxpool.Pool
}

func NewWalletRepo(pool *pgxpool.Pool) *WalletRepo {
	return &WalletRepo{pool: pool}
}

func (r *WalletRepo) ApplyOperation(ctx context.Context, id uuid.UUID, op domain.OperationType, amount int64) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var balance int64

	switch op {
	case domain.Deposit:
		err = tx.QueryRow(ctx, `
			INSERT INTO wallets (id, balance) VALUES ($1, $2)
			ON CONFLICT (id) DO UPDATE
			SET balance = wallets.balance + EXCLUDED.balance, updated_at = now()
			RETURNING balance`, id, amount).Scan(&balance)

	case domain.Withdraw:
		err = tx.QueryRow(ctx, `
			UPDATE wallets
			SET balance = balance - $2, updated_at = now()
			WHERE id = $1 AND balance >= $2
			RETURNING balance`, id, amount).Scan(&balance)

		if errors.Is(err, pgx.ErrNoRows) {
			var exists bool
			if err := tx.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM wallets WHERE id = $1)`, id).Scan(&exists); err != nil {
				return 0, fmt.Errorf("check wallet exists: %w", err)
			}
			if !exists {
				return 0, domain.ErrWalletNotFound
			}
			return 0, domain.ErrInsufficientFunds
		}

	default:
		return 0, domain.ErrInvalidOperation
	}

	if err != nil {
		return 0, fmt.Errorf("update balance: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO operations (wallet_id, operation_type, amount) VALUES ($1, $2, $3)`,
		id, string(op), amount)
	if err != nil {
		return 0, fmt.Errorf("insert operation: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return balance, nil
}

func (r *WalletRepo) GetBalance(ctx context.Context, id uuid.UUID) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx,
		`SELECT balance FROM wallets WHERE id = $1`, id).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.ErrWalletNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}
	return balance, nil
}
