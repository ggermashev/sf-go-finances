package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"sf-finances/src/types"
	"sf-finances/src/middlewares"
	"sf-finances/src/models"
	"sf-finances/src/services"
)

type AccountHandler struct {
	accountService *services.AccountService
	logger         *logrus.Logger
}

func NewAccountHandler(accountService *services.AccountService, logger *logrus.Logger) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		logger:         logger,
	}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		h.logger.Errorf("Ошибка получения userID: %v", err)
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	var req types.CreateAccountReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Ошибка декодирования: %v", err)
		http.Error(w, "Неверный формат", http.StatusBadRequest)
		return
	}

	if req.Currency != models.RUB {
		h.logger.Warnf("неподдерживаемая валюта: %s", req.Currency)
		http.Error(w, "Поддерживается только RUB", http.StatusBadRequest)
		return
	}

	newAccount, err := h.accountService.CreateAccount(r.Context(), userID, req.Currency)
	if err != nil {
		h.logger.Errorf("Не удалось создать счет: %v", err)
		http.Error(w, "Не удалось создать счет", http.StatusInternalServerError)
		return
	}

	resp := types.AccountRes{
		ID:        newAccount.ID,
		UserID:    newAccount.UserID,
		Balance:   newAccount.Balance,
		Currency:  newAccount.Currency,
		CreatedAt: newAccount.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("не удалось кодировать ответ: %v", err)
	}
}

func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		h.logger.Errorf("Не удалось получить UserId: %v", err)
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	accounts, err := h.accountService.GetAccountsByUserID(r.Context(), userID)
	if err != nil {
		h.logger.Errorf("Не удалось получить счета: %v", err)
		http.Error(w, "Не удалось получить счета", http.StatusInternalServerError)
		return
	}

	resp := types.AccountsListRes{
		Accounts: make([]types.AccountRes, 0, len(accounts)),
	}

	for _, acc := range accounts {
		resp.Accounts = append(resp.Accounts, types.AccountRes{
			ID:        acc.ID,
			UserID:    acc.UserID,
			Balance:   acc.Balance,
			Currency:  acc.Currency,
			CreatedAt: acc.CreatedAt.Format("2025-05-04T18:39:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("Ошибка кодирования: %v", err)
	}
}

func (h *AccountHandler) UpdateBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		h.logger.Errorf("Ошибка получения userID: %v", err)
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Warnf("Неверный ID счета: %v", err)
		http.Error(w, "Неверный ID счета", http.StatusBadRequest)
		return
	}

	var req types.UpdateBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Ошибка декодирования: %v", err)
		http.Error(w, "Неверный формат", http.StatusBadRequest)
		return
	}

	err = h.accountService.UpdateBalance(r.Context(), accountID, userID, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInsufficientFunds):
			h.logger.Warnf("Недостаточно средств: %v", err)
			http.Error(w, "Недостаточно средств", http.StatusBadRequest)
		default:
			h.logger.Errorf("Не удалось обновить баланс: %v", err)
			http.Error(w, "Не удалось обновить баланс", http.StatusInternalServerError)
		}
		return
	}

	updatedAccount, err := h.accountService.GetAccountByID(r.Context(), accountID, userID)
	if err != nil {
		h.logger.Errorf("Ошибка получения счета: %v", err)
		http.Error(w, "Не удалось получить данные счета", http.StatusInternalServerError)
		return
	}

	resp := types.AccountRes{
		ID:        updatedAccount.ID,
		UserID:    updatedAccount.UserID,
		Balance:   updatedAccount.Balance,
		Currency:  updatedAccount.Currency,
		CreatedAt: updatedAccount.CreatedAt.Format("2025-05-04T18:39:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("Ошибка кодирования: %v", err)
	}
}

func (h *AccountHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		h.logger.Errorf("Ошибка получения userID: %v", err)
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	var req types.TransferReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Ошибка декодирования: %v", err)
		http.Error(w, "Неверный формат", http.StatusBadRequest)
		return
	}

	err = h.accountService.Transfer(r.Context(), req.FromAccountID, req.ToAccountID, userID, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInsufficientFunds):
			h.logger.Warnf("Недостаточно средств: %v", err)
			http.Error(w, "Недостаточно средств", http.StatusBadRequest)
		case errors.Is(err, services.ErrSameAccount):
			h.logger.Warnf("Перевод на тот же счет: %v", err)
			http.Error(w, "Нельзя переводить на тот же счет", http.StatusBadRequest)
		case errors.Is(err, services.ErrNegativeAmount):
			h.logger.Warnf("перевод отрицательной суммы: %v", err)
			http.Error(w, "Сумма перевода должна быть положительной", http.StatusBadRequest)
		default:
			h.logger.Errorf("Ошибка перевода: %v", err)
			http.Error(w, "Не удалось перевести", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		h.logger.Errorf("Ошибка кодирования: %v", err)
	}
}

func (h *AccountHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		h.logger.Errorf("Ошибка получения userID: %v", err)
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Warnf("Неверный ID счета: %v", err)
		http.Error(w, "Неверный ID счета", http.StatusBadRequest)
		return
	}

	transactions, err := h.accountService.GetTransactionsByAccountID(r.Context(), accountID, userID)
	if err != nil {
		h.logger.Errorf("Не удалось получить транзакции: %v", err)
		http.Error(w, "Не удалось получить транзакции", http.StatusInternalServerError)
		return
	}

	resp := types.TransactionListRes{
		Transactions: make([]types.TransactionRes, 0, len(transactions)),
	}

	for _, tx := range transactions {
		resp.Transactions = append(resp.Transactions, types.TransactionRes{
			ID:        tx.ID,
			AccountID: tx.AccountID,
			Amount:    tx.Amount,
			Type:      tx.Type,
			Status:    tx.Status,
			CreatedAt: tx.CreatedAt.Format("2025-05-04T18:39:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("Ошибка кодирования: %v", err)
	}
}