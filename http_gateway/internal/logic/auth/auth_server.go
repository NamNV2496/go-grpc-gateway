package auth

import (
	"context"
	"log"

	"github.com/namnv2496/http_gateway/generated/pb"
	"github.com/namnv2496/http_gateway/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	jwtManager *JWTManager
	db         *gorm.DB
}

func NewAuthServer(
	jwtManner *JWTManager,
	db *gorm.DB,
) pb.AuthServiceServer {

	return &AuthServer{
		jwtManager: jwtManner,
		db:         db,
	}
}

func (auth *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {

	// GenPasswordTest()
	log.Println("Call login: ", req.Username)
	user, err := domain.GetUserByUserName(req.Username, auth.db)
	if err != nil {
		return nil, err
	}
	if !auth.jwtManager.IsCorrectPassword(req.Password, user.HashedPassword) {
		return nil, status.Errorf(codes.NotFound, "Cannot find username/password")
	}
	token, err := auth.jwtManager.Generate(&user)
	if err != nil {
		return nil, err
	}
	log.Println("Generated token: ", token)
	return &pb.LoginResponse{
		AccessToken: token,
	}, nil
}
