package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"sf-finances/src/models"
)

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) CreateAccount(ctx context.Context, userID int64, currency models.Currency) (*models.Account, error) {
	query := `
		INSERT INTO accounts (user_id, currency)
		VALUES ($1, $2)
		RETURNING id, user_id, balance, currency, created_at
	`
	var acc models.Account
	err := r.db.QueryRow(ctx, query, userID, currency).Scan(
		&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *AccountRepository) GetAccountByID(ctx context.Context, id int64) (*models.Account, error) {
	query := `
		SELECT id, user_id, balance, currency, created_at
		FROM accounts
		WHERE id = $1
	`
	var acc models.Account
	err := r.db.QueryRow(ctx, query, id).Scan(
		&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *AccountRepository) GetAccountsByUserID(ctx context.Context, userID int64) ([]*models.Account, error) {
	query := `
		SELECT id, user_id, balance, currency, created_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY id
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*models.Account
	for rows.Next() {
		var acc models.Account
		if err := rows.Scan(&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, &acc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *AccountRepository) UpdateBalance(ctx context.Context, id int64, amount decimal.Decimal) error {
	query := `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, amount, id)
	return err
}

func (r *AccountRepository) TransferBetweenAccounts(ctx context.Context, fromID, toID int64, amount decimal.Decimal) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Списание со счета отправителя
	updateFromQuery := `
		UPDATE accounts
		SET balance = balance - $1
		WHERE id = $2 AND balance >= $1
		RETURNING balance
	`
	var newBalance decimal.Decimal
	err = tx.QueryRow(ctx, updateFromQuery, amount, fromID).Scan(&newBalance)
	if err != nil {
		return err
	}

	// Пополнение счета получателя
	updateToQuery := `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
	`
	_, err = tx.Exec(ctx, updateToQuery, amount, toID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}