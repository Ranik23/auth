package main

import (
	"auth/internal/config"
	passwordservice "auth/internal/password-service"
	"auth/internal/server"
	"auth/internal/storage/postgres"
	pb "auth/proto/auth"
	pb2 "auth/proto/password"
	"context"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/dig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	container := dig.New()

	err := container.Provide(func() (net.Listener, error) {
		return net.Listen("tcp", ":8081")
	}, dig.Name("auth"))
	if err != nil {
		log.Fatalf("failed to register auth listener: %v", err)
	}

	err = container.Provide(func() (net.Listener, error) {
		return net.Listen("tcp", ":8082")
	}, dig.Name("password"))
	if err != nil {
		log.Fatalf("failed to register password listener: %v", err)
	}

	if err != nil {
		log.Fatalf("failed to register listener: %v", err)
	}

	err = container.Provide(func() (*config.Config, error) {
		return config.LoadConfig("config.yaml")
	})
	if err != nil {
		log.Fatalf("failed to register config: %v", err)
	}

	err = container.Provide(grpc.NewServer, dig.Name("auth"))
	if err != nil {
		log.Fatalf("failed to register gRPC server: %v", err)
	}

	err = container.Provide(grpc.NewServer, dig.Name("password"))
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
		return postgres.NewStoragePostgres(cfg)
	})
	if err != nil {
		log.Fatalf("failed to register storage: %v", err)
	}

	err = container.Provide(server.NewGRPCServer)
	if err != nil {
		log.Fatalf("failed to register gRPC server implementation: %v", err)
	}

	err = container.Provide(passwordservice.NewPasswordService)
	if err != nil {
		log.Fatalf("failed to register gRPC server implementation: %v", err)
	}

	err = container.Invoke(func(params struct {
		dig.In
		AuthListener     net.Listener `name:"auth"`
		PasswordListener net.Listener `name:"password"`
		GRPCServer       *grpc.Server `name:"auth"`
		GRPCServer2      *grpc.Server `name:"password"`
		GRPCServerImpl   *server.Server
		GRPCPasswordImpl *passwordservice.PasswordService
		Logger           *slog.Logger
	}) {
		pb.RegisterAuthServiceServer(params.GRPCServer, params.GRPCServerImpl)
		pb2.RegisterPasswordServiceServer(params.GRPCServer2, params.GRPCPasswordImpl)

		reflection.Register(params.GRPCServer)
		reflection.Register(params.GRPCServer2)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			params.Logger.Info("starting gRPC server", "address", params.AuthListener.Addr().String())
			if err := params.GRPCServer.Serve(params.AuthListener); err != nil {
				params.Logger.Error("failed to serve gRPC server", "error", err)
				cancel()
			}
		}()

		go func() {
			params.Logger.Info("starting gRPC server", "address", params.PasswordListener.Addr().String())
			if err := params.GRPCServer2.Serve(params.PasswordListener); err != nil {
				params.Logger.Error("failed to serve gRPC server", "error", err)
				cancel()
			}
		}()

		sigCH := make(chan os.Signal, 1)
		signal.Notify(sigCH, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-sigCH:
			params.Logger.Info("received shutdown signal")
		case <-ctx.Done():
			params.Logger.Info("shutting down due to error")
		}

		params.GRPCServer.GracefulStop()
		params.GRPCServer2.GracefulStop()
	})
	if err != nil {
		log.Fatalf("failed to invoke application: %v", err)
	}
}
