package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Polikarpov00/test_task_wallet/internal/domain"
	"github.com/Polikarpov00/test_task_wallet/internal/handler"
	"github.com/google/uuid"
)

type fakeService struct {
	balance int64
	err     error
	gotID   uuid.UUID
}

func (f *fakeService) ApplyOperation(_ context.Context, id uuid.UUID, _ domain.OperationType, _ int64) (int64, error) {
	f.gotID = id
	return f.balance, f.err
}

func (f *fakeService) GetBalance(_ context.Context, id uuid.UUID) (int64, error) {
	f.gotID = id
	return f.balance, f.err
}

func newMux(svc handler.WalletService) *http.ServeMux {
	mux := http.NewServeMux()
	handler.New(svc).Register(mux)
	return mux
}

// POST /api/v1/wallet с подделанным сервисом.
// Каждая доменная ошибка должна дать свой HTTP-код (400, 404, 422),
// невалидный JSON — 400, неизвестная ошибка — 500 без деталей
func TestApplyOperation(t *testing.T) {
	id := uuid.New()
	body := func(field string) string {
		return fmt.Sprintf(`{"%s":"%s","operationType":"DEPOSIT","amount":100}`, field, id)
	}

	tests := []struct {
		name        string
		body        string
		svcErr      error
		wantStatus  int
		wantBody    string
		notWantBody string
	}{
		{"ok", body("walletId"), nil, http.StatusOK, `"balance":500`, ""},
		{"invalid json", `{bad json`, nil, http.StatusBadRequest, "invalid request body", ""},
		{"validation error", body("walletId"), domain.ErrInvalidAmount, http.StatusBadRequest, domain.ErrInvalidAmount.Error(), ""},
		{"wallet not found", body("walletId"), domain.ErrWalletNotFound, http.StatusNotFound, domain.ErrWalletNotFound.Error(), ""},
		{"insufficient funds", body("walletId"), domain.ErrInsufficientFunds, http.StatusUnprocessableEntity, domain.ErrInsufficientFunds.Error(), ""},
		{"internal error hidden", body("walletId"), errors.New("dial tcp 10.0.0.5:5432: connection refused"),
			http.StatusInternalServerError, "internal error", "10.0.0.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{balance: 500, err: tt.svcErr}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			newMux(svc).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %s, want it to contain %q", rec.Body, tt.wantBody)
			}
			if tt.notWantBody != "" && strings.Contains(rec.Body.String(), tt.notWantBody) {
				t.Errorf("body leaks internal details: %s", rec.Body)
			}
			if tt.wantStatus == http.StatusOK && svc.gotID != id {
				t.Errorf("service got id %s, want %s", svc.gotID, id)
			}
		})
	}
}

// GET /api/v1/wallets/{id}.
// Валидный id — 200 и баланс, невалидный UUID — 400 до вызова сервиса,
// несуществующий кошелёк — 404.
func TestGetBalance(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name       string
		path       string
		svcErr     error
		wantStatus int
		wantBody   string
	}{
		{"ok", "/api/v1/wallets/" + id.String(), nil, http.StatusOK, `"balance":500`},
		{"invalid uuid", "/api/v1/wallets/abc", nil, http.StatusBadRequest, "invalid wallet id"},
		{"not found", "/api/v1/wallets/" + id.String(), domain.ErrWalletNotFound, http.StatusNotFound, domain.ErrWalletNotFound.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{balance: 500, err: tt.svcErr}

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			newMux(svc).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %s, want it to contain %q", rec.Body, tt.wantBody)
			}
		})
	}
}
