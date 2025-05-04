package types;

import (
	"github.com/shopspring/decimal"
	"sf-finances/src/models"
)

type CreateAccountReq struct {
	Currency models.Currency `json:"currency"`
}

type UpdateBalanceRequest struct {
	Amount decimal.Decimal `json:"amount"`
}

type TransferReq struct {
	FromAccountID int64           `json:"from_account_id"`
	ToAccountID   int64           `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
}

type AccountRes struct {
	ID        int64            `json:"id"`
	UserID    int64            `json:"user_id"`
	Balance   decimal.Decimal  `json:"balance"`
	Currency  models.Currency `json:"currency"`
	CreatedAt string           `json:"created_at"`
}

type TransactionRes struct {
	ID        int64              `json:"id"`
	AccountID int64              `json:"account_id"`
	Amount    decimal.Decimal    `json:"amount"`
	Type      models.TransactionType   `json:"type"`
	Status    models.TransactionStatus `json:"status"`
	CreatedAt string             `json:"created_at"`
}

type AccountsListRes struct {
	Accounts []AccountRes `json:"accounts"`
}

type TransactionListRes struct {
	Transactions []TransactionRes `json:"transactions"`
}