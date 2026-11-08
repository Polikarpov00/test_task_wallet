package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Polikarpov00/test_task_wallet/internal/domain"
	"github.com/google/uuid"
)

type WalletService interface {
	ApplyOperation(ctx context.Context, id uuid.UUID, op domain.OperationType, amount int64) (int64, error)
	GetBalance(ctx context.Context, id uuid.UUID) (int64, error)
}

type Handler struct {
	svc WalletService
}

func New(svc WalletService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/wallet", h.applyOperation)
	mux.HandleFunc("GET /api/v1/wallets/{id}", h.getBalance)
}

type operationRequest struct {
	WalletID      uuid.UUID            `json:"walletId"`
	OperationType domain.OperationType `json:"operationType"`
	Amount        int64                `json:"amount"`
}

type walletResponse struct {
	WalletID uuid.UUID `json:"walletId"`
	Balance  int64     `json:"balance"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) applyOperation(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req operationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	balance, err := h.svc.ApplyOperation(r.Context(), req.WalletID, req.OperationType, req.Amount)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, walletResponse{WalletID: req.WalletID, Balance: balance})
}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid wallet id"})
		return
	}

	balance, err := h.svc.GetBalance(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, walletResponse{WalletID: id, Balance: balance})
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidWalletID),
		errors.Is(err, domain.ErrInvalidOperation),
		errors.Is(err, domain.ErrInvalidAmount):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrWalletNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInsufficientFunds):
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
	default:
		log.Printf("internal error: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"}) // для безопасности, чтобы не отдавать лишние данные типа айпишника
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}
