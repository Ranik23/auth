package passwordservice

import (
	"auth/internal/config"
	pb "auth/proto/password"
	"context"
	"log"
)

type PasswordService struct {
	pb.UnimplementedPasswordServiceServer
	cfg *config.Config
}

func NewPasswordService(cfg *config.Config) *PasswordService {
	return &PasswordService{
		cfg: cfg,
	}
}

func (ps *PasswordService) ChangePassword(ctx context.Context, req *pb.ChangePassworsRequest) (*pb.ChangePasswordResponse, error) {
	log.Println("password successfully changed")
	return nil, nil
}

