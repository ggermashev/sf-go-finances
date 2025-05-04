package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"sf-finances/src/models"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) CreateTransaction(ctx context.Context, accountID int64, amount decimal.Decimal,
	txType models.TransactionType, status models.TransactionStatus) (*models.Transaction, error) {
	query := `
		INSERT INTO transactions
		VALUES ($1, $2, $3, $4)
		RETURNING id, account_id, amount, type, status, created_at
	`
	var tx models.Transaction
	err := r.db.QueryRow(ctx, query, accountID, amount, txType, status).Scan(
		&tx.ID, &tx.AccountID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *TransactionRepository) GetTransactionsByAccountID(ctx context.Context, accountID int64) ([]*models.Transaction, error) {
	query := `
		SELECT id, account_id, amount, type, status, created_at
		FROM transactions
		WHERE account_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(&tx.ID, &tx.AccountID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *TransactionRepository) GetTransactionsByUserID(ctx context.Context, userID int64) ([]*models.Transaction, error) {
	query := `
		SELECT t.id, t.account_id, t.amount, t.type, t.status, t.created_at
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE a.user_id = $1
		ORDER BY t.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(&tx.ID, &tx.AccountID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}