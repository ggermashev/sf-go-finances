package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
	"sf-finances/src/types"
	"sf-finances/src/services"
)

type AuthHandler struct {
	userService services.UserService
	logger      *logrus.Logger
}

func NewAuthHandler(userService services.UserService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req types.RegisterReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Warn("Ошибка декодирования")
		http.Error(w, "Неверный формат", http.StatusBadRequest)
		return
	}

	userID, err := h.userService.Register(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Warn("Ошибка при регистрации")

		if errors.Is(err, services.ErrUserExists) {
			http.Error(w, "Пользователь с таким логином уже существует", http.StatusConflict)
			return
		}

		http.Error(w, "Ошибка при регистрации", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"message": "Пользователь зарегистрирован",
		"user_id": userID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithError(err).Error("Ошибка формирования ответа")
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req types.LoginReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Warn("Ошибка декодирования")
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email и пароль обязательны", http.StatusBadRequest)
		return
	}

	token, err := h.userService.Login(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Warn("Ошибка при авторизации")

		if errors.Is(err, services.ErrInvalidCredentials) {
			http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
			return
		}

		http.Error(w, "Ошибка авторизации", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := types.LoginRes{Token: token}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithError(err).Error("Ошибка формирования ответа")
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
}