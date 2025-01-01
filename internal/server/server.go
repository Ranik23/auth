package server

import (
	"auth/internal/storage/postgres"
	pb "auth/proto/auth"
	pb2 "auth/proto/password"
	"context"
	"log/slog"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	secret = "your-secret-key"
)

type Server struct {
	pb.UnimplementedAuthServiceServer
	storage 			postgres.Storage
	logger 				*slog.Logger
}

func NewGRPCServer(database postgres.Storage, logger *slog.Logger) *Server{
	return &Server{
		storage: database,
		logger: logger,
	}
}


func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {

	if !govalidator.IsEmail(req.Email) {
		return &pb.RegisterResponse{Message: "email is not vaild"}, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash the password")
		return &pb.RegisterResponse{Message: "failed"}, err
	}
	if err := s.storage.SaveUser(req.Username, req.Email, req.Age, hashedPassword); err != nil {
		s.logger.Error("failed to save the user")
		return &pb.RegisterResponse{Message: "failed"}, err
	}
	return &pb.RegisterResponse{Message: "succesfully registered"}, nil
}


func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {

	username := req.GetUsername()
	password := req.GetPassword()

	user, err := s.storage.GetUser(username)
	if err != nil {
		s.logger.Error("failed to get the user")
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password)); err != nil {
		s.logger.Error("failed to compare hash and password")
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.UserName,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return &pb.LoginResponse{Token: tokenString}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Errorf(codes.Internal, "unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}


	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &pb.ValidateTokenResponse{Valid: true, Username: claims["username"].(string)}, nil
	}

	return &pb.ValidateTokenResponse{Valid: false}, nil
}

func (s *Server) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	

	connection, err := grpc.NewClient("localhost:8082", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return &pb.ResetPasswordResponse{
			Message: "failed to connect to the server",
			Password: "none",
		}, nil
	}

	defer connection.Close()


	client := pb2.NewPasswordServiceClient(connection)

	resp, err := client.ResetPassword(ctx, &pb2.ResetPassworsRequest{
		Email : req.GetEmail(),
	})
	if err != nil {
		return &pb.ResetPasswordResponse{
			Message: "failed to reset the password:",
			Password: "error",
		}, nil
	}

	//newPassword := resp.Password

	return &pb.ResetPasswordResponse{
		Message: "successfully reset the password",
		Password: resp.Password,
	}, nil

}