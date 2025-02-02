package authservice

import (
	"auth/internal/storage/postgres"
	pb "auth/proto/auth"
	pb2 "auth/proto/password"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	secret           = "your-secret-key" // Вынесите в конфигурацию
	tokenExpiration  = 72 * time.Hour
	passwordMinLength = 8
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	storage postgres.Storage
	logger  *slog.Logger
}

func NewGRPCServer(database postgres.Storage, logger *slog.Logger) *AuthService {
	return &AuthService{
		storage: database,
		logger:  logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Валидация email
	if !govalidator.IsEmail(req.Email) {
		s.logger.Warn("invalid email", "email", req.Email)
		return &pb.RegisterResponse{Message: "email is not valid"}, nil
	}

	// Валидация пароля
	if len(req.Password) < passwordMinLength {
		s.logger.Warn("password is too short", "length", len(req.Password))
		return &pb.RegisterResponse{Message: "password must be at least 8 characters long"}, nil
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return &pb.RegisterResponse{Message: "internal server error"}, status.Errorf(codes.Internal, "failed to hash password")
	}

	// Сохранение пользователя
	if err := s.storage.SaveUser(req.Username, req.Email, req.Age, hashedPassword); err != nil {
		s.logger.Error("failed to save user", "error", err)
		return &pb.RegisterResponse{Message: "internal server error"}, status.Errorf(codes.Internal, "failed to save user")
	}

	s.logger.Info("user registered successfully", "username", req.Username)
	return &pb.RegisterResponse{Message: "successfully registered"}, nil
}

func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	username := req.GetUsername()
	password := req.GetPassword()

	// Получение пользователя
	user, err := s.storage.GetUserByUserName(username)
	if err != nil {
		s.logger.Error("failed to get user", "username", username, "error", err)
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password)); err != nil {
		s.logger.Warn("invalid password", "username", username)
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}

	// Генерация JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.UserName,
		"exp":      time.Now().Add(tokenExpiration).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		s.logger.Error("failed to generate token", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	s.logger.Info("user logged in successfully", "username", username)
	return &pb.LoginResponse{Token: tokenString}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		s.logger.Warn("invalid token", "error", err)
		return &pb.ValidateTokenResponse{Valid: false}, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username := claims["username"].(string)
		s.logger.Info("token validated successfully", "username", username)
		return &pb.ValidateTokenResponse{Valid: true, Username: username}, nil
	}

	s.logger.Warn("invalid token claims")
	return &pb.ValidateTokenResponse{Valid: false}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	conn, err := grpc.NewClient("localhost:8082", grpc.WithTransportCredentials(insecure.NewCredentials())) // из редиса брать порты и хосты
	if err != nil {
		s.logger.Error("failed to connect to password service", "error", err)
		return &pb.ResetPasswordResponse{Message: "internal server error"}, status.Errorf(codes.Internal, "failed to connect to password service")
	}
	defer conn.Close()

	client := pb2.NewPasswordServiceClient(conn)
	
	_, err = client.ChangePassword(ctx, &pb2.ChangePasswordRequest{
		Email:       req.GetEmail(),
		NewPassword: req.GetNewPassword(),
	})
	if err != nil {
		s.logger.Error("failed to reset password", "error", err)
		return &pb.ResetPasswordResponse{Message: "internal server error"}, status.Errorf(codes.Internal, "failed to reset password")
	}

	s.logger.Info("password reset successfully", "email", req.Email)
	return &pb.ResetPasswordResponse{Message: "successfully reset the password"}, nil
}