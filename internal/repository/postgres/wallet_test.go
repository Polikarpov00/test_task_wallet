package postgres

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Polikarpov00/test_task_wallet/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Подключение к тестовой БД
func setupRepo(t *testing.T) *WalletRepo {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://wallet:wallet@localhost:5433/wallet?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewWalletRepo(pool)
}

// Параллельно 100 депозитов и 100 снятий по 10 на одном кошельке.
// Если хоть одно обновление потеряется, итоговый баланс не сойдётся.
func TestApplyOperation_ConcurrentMixed(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()
	id := uuid.New()

	const initial = 1000
	if _, err := repo.ApplyOperation(ctx, id, domain.Deposit, initial); err != nil {
		t.Fatalf("initial deposit: %v", err)
	}

	const n = 200
	var wg sync.WaitGroup
	var failed atomic.Int64

	for i := range n {
		op := domain.Withdraw
		if i%2 == 0 {
			op = domain.Deposit
		}
		wg.Go(func() {
			if _, err := repo.ApplyOperation(ctx, id, op, 10); err != nil {
				t.Errorf("op %s: %v", op, err)
				failed.Add(1)
			}
		})
	}
	wg.Wait()

	balance, err := repo.GetBalance(ctx, id)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance != initial {
		t.Errorf("balance = %d, want %d (failed ops: %d)", balance, initial, failed.Load())
	}
}

// Баланс 100,
// 50 горутин снимают по 10.
// Ожидаем 10 снятий, остальные 40 с ошибкой ErrInsufficientFunds,
// баланс = 0 и не опустился ниже 0.
func TestApplyOperation_ConcurrentOverdraw(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()
	id := uuid.New()

	if _, err := repo.ApplyOperation(ctx, id, domain.Deposit, 100); err != nil {
		t.Fatalf("initial deposit: %v", err)
	}

	const n = 50
	var wg sync.WaitGroup
	var ok, insufficient atomic.Int64

	for range n {
		wg.Go(func() {
			_, err := repo.ApplyOperation(ctx, id, domain.Withdraw, 10)
			switch {
			case err == nil:
				ok.Add(1)
			case errors.Is(err, domain.ErrInsufficientFunds):
				insufficient.Add(1)
			default:
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
	wg.Wait()

	if ok.Load() != 10 || insufficient.Load() != 40 {
		t.Errorf("ok = %d, insufficient = %d, want 10 and 40", ok.Load(), insufficient.Load())
	}

	balance, err := repo.GetBalance(ctx, id)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance != 0 {
		t.Errorf("balance = %d, want 0", balance)
	}
}
