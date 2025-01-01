package passwordservice

import (
	"auth/internal/config"
	pb "auth/proto/password"
	"context"
	"os"
	"github.com/icrowley/fake"
	"gopkg.in/gomail.v2"
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

func (ps *PasswordService) ResetPassword(ctx context.Context, req *pb.ResetPassworsRequest) (*pb.ResetPasswordResponse, error) {
	m := gomail.NewMessage()

    m.SetHeader("From", "antonfedorov190@ayndex.ru")
    m.SetHeader("To", req.GetEmail()) 
    m.SetHeader("Subject", "Hello!") 
    m.SetBody("text/plain", "This is a test email.")

	password := os.Getenv("EMAIL_PASSWORD")

	dialer := gomail.NewDialer("smtp.yandex.ru", 465, ps.cfg.SMTPConfig.Email, password)

	if err := dialer.DialAndSend(m); err != nil {
		return &pb.ResetPasswordResponse{
			Messagge: "failed to reset the password",
			Password: "none",
		}, err
	}

	return &pb.ResetPasswordResponse{
		Messagge: "successfully resetted",
		Password: fake.Password(5, 20, true, true, true),
	}, nil
} 

