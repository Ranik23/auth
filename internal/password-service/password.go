package passwordservice

import (
	"auth/internal/config"
	"auth/internal/storage/postgres"
	pb "auth/proto/password"
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"golang.org/x/crypto/bcrypt"
)

type PasswordService struct {
	pb.UnimplementedPasswordServiceServer
	cfg         *config.Config
	storage     postgres.Storage
	kafkaWriter *kafka.Writer
}

func NewPasswordService(cfg *config.Config, strg postgres.Storage, kafkaWriter *kafka.Writer) *PasswordService {
	return &PasswordService{
		cfg:         cfg,
		storage:     strg,
		kafkaWriter: kafkaWriter,
	}
}

func (ps *PasswordService) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	newPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetNewPassword()), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to generate password hash: %v", err)
		return &pb.ChangePasswordResponse{
			Message: "internal server error",
		}, err
	}

	if err := ps.storage.ChangePassword(req.GetEmail(), newPassword); err != nil {
		log.Printf("Failed to change password in storage: %v", err)
		return &pb.ChangePasswordResponse{
			Message: "internal server error",
		}, err
	}

	log.Println("Password successfully changed")

	if err := ps.sendNotificationEvent(req.GetEmail(), "password successfully changed"); err != nil {
		log.Printf("Failed to send notification: %v", err)
	}

	return &pb.ChangePasswordResponse{
		Message: "password successfully changed",
	}, nil
}

func (ps *PasswordService) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	user, err := ps.storage.GetUserByEmail(req.GetEmail())
	if err != nil {
		log.Printf("Failed to get user by email: %v", err)
		return &pb.UpdatePasswordResponse{
			Message: "internal server error",
		}, err
	}

	if err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(req.GetOldPassword())); err != nil {
		log.Printf("Old password is incorrect for user %s", req.GetEmail())
		return &pb.UpdatePasswordResponse{
			Message: "old password is incorrect",
		}, nil
	}

	// Используем метод ChangePassword сервиса для обновления пароля
	resp, err := ps.ChangePassword(ctx, &pb.ChangePasswordRequest{
		Email:       req.GetEmail(),
		NewPassword: req.GetNewPassword(),
	})
	if err != nil {
		log.Printf("Failed to change password: %v", err)
		return &pb.UpdatePasswordResponse{
			Message: "internal server error",
		}, err
	}

	return &pb.UpdatePasswordResponse{
		Message: resp.Message,
	}, nil
}

func (ps *PasswordService) sendNotificationEvent(email, message string) error {
	event := map[string]string{
		"email":   email,
		"message": message,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return err
	}

	if err := ps.kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Value: eventJSON,
	}); err != nil {
		log.Printf("Failed to send Kafka message: %v", err)
		return err
	}

	return nil
}