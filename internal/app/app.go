package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stack/internal/config"
	deliveryHttp "stack/internal/delivery/http"
	"stack/internal/repository/postgres"
	"stack/internal/usecase"
	"stack/pkg/hasher"
	"stack/pkg/token"
)

func Run() error {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Инициализация подключения к БД
	dbpool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}
	defer dbpool.Close()
	log.Println("Успешное подключение к PostgreSQL")

	// 2. Инициализация инфраструктурных компонентов безопасности
	pwdHasher := hasher.NewBcryptHasher(0)
	tokenManager := token.NewJWTManager(cfg.JWTSecret)

	// 3. Инициализация репозиториев
	userRepo := postgres.NewUserRepository(dbpool)

	// 4. Инициализация сценариев использования (Use Cases)
	authUsecase := usecase.NewAuthUsecase(userRepo, pwdHasher, tokenManager, cfg.TokenTTL)

	// 5. Инициализация HTTP обработчиков и маршрутизатора
	authHandler := deliveryHttp.NewAuthHandler(authUsecase)
	router := deliveryHttp.NewRouter(authHandler)

	// 6. Конфигурация HTTP сервера
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Канал для перехвата сигналов завершения работы (graceful shutdown)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	serverErrChan := make(chan error, 1)
	go func() {
		log.Printf("HTTP сервер запущен на порту :%s\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	select {
	case err := <-serverErrChan:
		return fmt.Errorf("ошибка HTTP сервера: %w", err)
	case sig := <-shutdownChan:
		log.Printf("Получен сигнал %v, плавное завершение работы...", sig)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("ошибка при graceful shutdown: %w", err)
		}
		log.Println("Сервер успешно остановлен")
	}

	return nil
}
