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
	cfg *config.Config
	storage postgres.Storage
	kafkaWriter *kafka.Writer
}

func NewPasswordService(cfg *config.Config, strg postgres.Storage, kafkaWriter *kafka.Writer) *PasswordService {
	return &PasswordService{
		cfg: cfg,
		storage: strg,
		kafkaWriter: kafkaWriter,
	}
}

func (ps *PasswordService) ChangePassword(ctx context.Context, req *pb.ChangePassworsRequest) (*pb.ChangePasswordResponse, error) {

	newPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetNewPassword()), bcrypt.DefaultCost)
	if err != nil {
		return &pb.ChangePasswordResponse{
			Messagge: "internal server error",
		}, nil
	}
	if err := ps.storage.ChangePassword(req.GetEmail(), newPassword); err != nil {
		return &pb.ChangePasswordResponse{
			Messagge : "internal server error",
		}, nil
	}

	log.Println("password successfully changed")
	return nil, nil
}


func (ps *PasswordService) sendNotificationEvent(email, message string) error {
	event := map[string]string {
		"email": email,
		"message": message,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Println("failed to marshall event")
		return err
	}

	return ps.kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Value: eventJSON, 
	})
}

