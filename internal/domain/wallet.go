package domain

import (
	"errors"

	"github.com/google/uuid"
)

type OperationType string

const (
	Deposit  OperationType = "DEPOSIT"
	Withdraw OperationType = "WITHDRAW"
)

func (t OperationType) Valid() bool {
	return t == Deposit || t == Withdraw
}

var (
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidOperation  = errors.New("invalid operation type")
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrInvalidWalletID   = errors.New("wallet id is required")
)

type Wallet struct {
	ID      uuid.UUID
	Balance int64
}
