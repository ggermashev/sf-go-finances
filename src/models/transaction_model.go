package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type TransactionStatus string
const (
	PENDING   TransactionStatus = "PENDING"
	COMPLETED TransactionStatus = "COMPLETED"
	FAILED    TransactionStatus = "FAILED"
)

type TransactionType string
const (
	DEPOSIT    TransactionType = "DEPOSIT"
	WITHDRAWAL TransactionType = "WITHDRAWAL"
	TRANSFER   TransactionType = "TRANSFER"
)

type Transaction struct {
	ID        int64           `db:"id"          json:"id"`
	AccountID int64           `db:"account_id"  json:"account_id"`
	Amount    decimal.Decimal `db:"amount" json:"amount"`
	Type      TransactionType            `db:"type"        json:"type"`
	Status    TransactionStatus          `db:"status"      json:"status"`
	CreatedAt time.Time       `db:"created_at"  json:"created_at"`
}