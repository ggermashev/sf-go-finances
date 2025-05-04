package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"sf-finances/src/config"
	"sf-finances/src/handlers"
	"sf-finances/src/middlewares"
	"sf-finances/src/repository"
	"sf-finances/src/services"
)


func main() {
	// Инициализация логера
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Загрузка конфигурации
	ctx := context.Background()
	dbCfg := config.GetDBConfig()
	jwtCfg := config.GetJWTConfig()
	cryptoCfg := config.GetCryptoConfig()

	pool, err := config.CreatePgPool(ctx, dbCfg)
	if err != nil {
		logger.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer pool.Close()
	logger.Info("Подключено к БД")

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(pool)
	accountRepo := repository.NewAccountRepository(pool)
	transactionRepo := repository.NewTransactionRepository(pool)
	cardRepo := repository.NewCardRepository(pool)

	// Инициализация сервисов
	authService := services.NewAuthService(userRepo, jwtCfg)
	accountService := services.NewAccountService(accountRepo, transactionRepo)
	cardService := services.NewCardService(cardRepo, pool, cryptoCfg.HMACKey)

	// Инициализация обработчиков
	authHandler := handler.NewAuthHandler(authService, logger)
	accountHandler := handler.NewAccountHandler(accountService, logger)
	cardHandler := handler.NewCardHandler(cardService, logger)

	// JWT middleware
	jwtMiddleware := middlewares.NewJWTMiddleware(authService, logger)

	// Настройка маршрутизатора
	r := mux.NewRouter().PathPrefix("/api").Subrouter()

	// Публичные маршруты (без аутентификации)
	r.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
	r.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)

	// Защищенные маршруты (с проверкой JWT)
	apiRouter := r.PathPrefix("").Subrouter()
	apiRouter.Use(jwtMiddleware.Middleware)

	// Маршруты для счетов
	apiRouter.HandleFunc("/accounts", accountHandler.CreateAccount).Methods(http.MethodPost)
	apiRouter.HandleFunc("/accounts", accountHandler.GetAccounts).Methods(http.MethodGet)
	apiRouter.HandleFunc("/accounts/{id}/balance", accountHandler.UpdateBalance).Methods(http.MethodPatch)
	apiRouter.HandleFunc("/accounts/{id}/transactions", accountHandler.GetTransactions).Methods(http.MethodGet)
	apiRouter.HandleFunc("/transfer", accountHandler.Transfer).Methods(http.MethodPost)

	// Маршруты для карт
	apiRouter.HandleFunc("/cards", cardHandler.CreateCard).Methods(http.MethodPost)
	apiRouter.HandleFunc("/cards", cardHandler.GetCards).Methods(http.MethodGet)
	apiRouter.HandleFunc("/cards/{id}", cardHandler.GetCardDetails).Methods(http.MethodGet)
	apiRouter.HandleFunc("/payments", cardHandler.ProcessPayment).Methods(http.MethodPost)

	// Настройка сервера
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", "8080"),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера в горутине
	go func() {
		logger.Infof("Сервер на порту %s", "8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Канал для сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Завершение сервера")

	// Ожидание завершения текущих запросов
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Fatalf("Ошибка остановки сервера: %v", err)
	}
	logger.Info("Сервер остановлен")
}