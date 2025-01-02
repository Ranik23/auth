package main

import (
	"auth/internal/config"
	passwordservice "auth/internal/password-service"
	authservice "auth/internal/auth-service"
	"auth/internal/storage/postgres"
	pb "auth/proto/auth"
	pb2 "auth/proto/password"
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Инициализация логгера
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Инициализация хранилища
	storage, err := postgres.NewStoragePostgres(cfg)
	if err != nil {
		logger.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}

	// Инициализация gRPC серверов
	authServer := grpc.NewServer()
	passwordServer := grpc.NewServer()

	// Инициализация сервисов
	authService := authservice.NewGRPCServer(storage, logger)
	passwordService := passwordservice.NewPasswordService(cfg, storage, nil) // Kafka writer можно добавить позже

	// Регистрация сервисов на gRPC серверах
	pb.RegisterAuthServiceServer(authServer, authService)
	pb2.RegisterPasswordServiceServer(passwordServer, passwordService)

	// Включение reflection для отладки
	reflection.Register(authServer)
	reflection.Register(passwordServer)

	// Создание слушателей
	authListener, err := net.Listen("tcp", ":8081")
	if err != nil {
		logger.Error("failed to create auth listener", "error", err)
		os.Exit(1)
	}

	passwordListener, err := net.Listen("tcp", ":8082")
	if err != nil {
		logger.Error("failed to create password listener", "error", err)
		os.Exit(1)
	}

	// Контекст для управления жизненным циклом серверов
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск gRPC серверов
	go startGRPCServer(ctx, authServer, authListener, logger, "auth", cancel)
	go startGRPCServer(ctx, passwordServer, passwordListener, logger, "password", cancel)

	// Обработка сигналов завершения
	handleShutdownSignals(ctx, cancel, logger)

	// Грациозное завершение работы серверов
	authServer.GracefulStop()
	passwordServer.GracefulStop()
}

// startGRPCServer запускает gRPC сервер
func startGRPCServer(ctx context.Context, server *grpc.Server, listener net.Listener, logger *slog.Logger, name string, cancel context.CancelFunc) {
	logger.Info("starting gRPC server", "name", name, "address", listener.Addr().String())
	if err := server.Serve(listener); err != nil {
		logger.Error("failed to serve gRPC server", "name", name, "error", err)
		cancelFromContext(cancel) // Передаем cancel функцию
	}
}

// handleShutdownSignals обрабатывает сигналы завершения
func handleShutdownSignals(ctx context.Context, cancel context.CancelFunc, logger *slog.Logger) {
	sigCH := make(chan os.Signal, 1)
	signal.Notify(sigCH, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCH:
		logger.Info("received shutdown signal")
	case <-ctx.Done():
		logger.Info("shutting down due to error")
	}
}

// cancelFromContext вызывает cancel функцию
func cancelFromContext(cancel context.CancelFunc) {
	if cancel != nil {
		cancel()
	}
}