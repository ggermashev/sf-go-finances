package services

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"sf-finances/src/models"
	"sf-finances/src/repository"
)

var (
	ErrInsufficientFunds = errors.New("не хватает средств")
	ErrSameAccount       = errors.New("нельзя делать перевод на тот же счет")
	ErrNegativeAmount    = errors.New("сумма должна быть положительной")
)

type AccountService struct {
	accountRepo     *repository.AccountRepository
	transactionRepo *repository.TransactionRepository
}

func NewAccountService(accountRepo *repository.AccountRepository, transactionRepo *repository.TransactionRepository) *AccountService {
	return &AccountService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, userID int64, currency models.Currency) (*models.Account, error) {
	return s.accountRepo.CreateAccount(ctx, userID, currency)
}

func (s *AccountService) GetAccountByID(ctx context.Context, id int64, userID int64) (*models.Account, error) {
	acc, err := s.accountRepo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if acc.UserID != userID {
		return nil, errors.New("счет принадлежит другому пользователю")
	}

	return acc, nil
}

func (s *AccountService) GetAccountsByUserID(ctx context.Context, userID int64) ([]*models.Account, error) {
	return s.accountRepo.GetAccountsByUserID(ctx, userID)
}

func (s *AccountService) UpdateBalance(ctx context.Context, id int64, userID int64, amount decimal.Decimal) error {
	if amount.Equal(decimal.Zero) {
		return errors.New("сумма не указана")
	}

	acc, err := s.GetAccountByID(ctx, id, userID)
	if err != nil {
		return err
	}

	if amount.LessThan(decimal.Zero) && acc.Balance.Add(amount).LessThan(decimal.Zero) {
		return ErrInsufficientFunds
	}

	txType := models.WITHDRAWAL
	if amount.GreaterThan(decimal.Zero) {
		txType = models.DEPOSIT
	}

	err = s.accountRepo.UpdateBalance(ctx, id, amount)
	if err != nil {
		return err
	}

	absAmount := amount.Abs()
	_, err = s.transactionRepo.CreateTransaction(ctx, id, absAmount, txType, models.COMPLETED)

	return err
}

func (s *AccountService) Transfer(ctx context.Context, fromID, toID int64, userID int64, amount decimal.Decimal) error {
	if fromID == toID {
		return ErrSameAccount
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return ErrNegativeAmount
	}

	fromAcc, err := s.GetAccountByID(ctx, fromID, userID)
	if err != nil {
		return err
	}

	if fromAcc.Balance.LessThan(amount) {
		return ErrInsufficientFunds
	}

	_, err = s.accountRepo.GetAccountByID(ctx, toID)
	if err != nil {
		return err
	}

	err = s.accountRepo.TransferBetweenAccounts(ctx, fromID, toID, amount)
	if err != nil {
		return err
	}

	_, err = s.transactionRepo.CreateTransaction(ctx, fromID, amount, models.DEPOSIT, models.COMPLETED)
	if err != nil {
		return err
	}

	_, err = s.transactionRepo.CreateTransaction(ctx, toID, amount, models.WITHDRAWAL, models.COMPLETED)
	return err
}

func (s *AccountService) GetTransactionsByAccountID(ctx context.Context, accountID int64, userID int64) ([]*models.Transaction, error) {
	_, err := s.GetAccountByID(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}

	return s.transactionRepo.GetTransactionsByAccountID(ctx, accountID)
}

func (s *AccountService) GetTransactionsByUserID(ctx context.Context, userID int64) ([]*models.Transaction, error) {
	return s.transactionRepo.GetTransactionsByUserID(ctx, userID)
}