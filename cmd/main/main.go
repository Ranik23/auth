package main

import (
	"auth/internal/config"
	"auth/internal/server"
	"auth/internal/storage/postgres"
	pb "auth/proto"
	"log"
	"log/slog"
	"net"
	"os"

	"go.uber.org/dig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	container := dig.New()

	err := container.Provide(func() (net.Listener, error) {
		return net.Listen("tcp", ":8081")
	})
	if err != nil {
		log.Fatalf("failed to register listener: %v", err)
	}

	err = container.Provide(func() (*config.Config, error) {
		return config.LoadConfig("config.yaml")
	})
	if err != nil {
		log.Fatalf("failed to register config: %v", err)
	}

	err = container.Provide(grpc.NewServer)
	if err != nil {
		log.Fatalf("failed to register gRPC server: %v", err)
	}

	err = container.Provide(func() *slog.Logger {
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	})
	if err != nil {
		log.Fatalf("failed to register logger: %v", err)
	}

	err = container.Provide(func(cfg *config.Config) (postgres.Storage, error) {
		return postgres.NewStoragePostgres(cfg) // могу заменить на любое дерьмо, возвращаем мы интерфейс. каждый другой компонент
		// контейнера зависит только от интерфейса. контейнер оперирует только интерфейсом
	})
	if err != nil {
		log.Fatalf("failed to register storage: %v", err)
	}

	err = container.Provide(server.NewGRPCServer)
	if err != nil {
		log.Fatalf("failed to register gRPC server implementation: %v", err)
	}

	err = container.Invoke(func(
		listener net.Listener,
		grpcServer *grpc.Server,
		grpcImpl *server.Server,
		logger *slog.Logger,
	) {

		pb.RegisterAuthServiceServer(grpcServer, grpcImpl)
		reflection.Register(grpcServer)

		logger.Info("starting gRPC server", "address", listener.Addr().String())
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error("failed to serve gRPC server", "error", err)
			os.Exit(1)
		}
	})
	if err != nil {
		log.Fatalf("failed to invoke application: %v", err)
	}
}