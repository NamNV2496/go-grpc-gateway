package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/namnv2496/http_gateway/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type UserClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Role     string `json:"role"`
}

func NewJWTManager(
	secretKey string,
	tokenDuration time.Duration,
) *JWTManager {
	return &JWTManager{secretKey, tokenDuration}
}

// func GenPasswordTest() {
// 	hashPassword, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
// 	if err != nil {
// 		return
// 	}
// 	log.Println(string(hashPassword))
// }

func (jwtManner *JWTManager) Generate(user *domain.User) (string, error) {

	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(jwtManner.tokenDuration).Unix(),
		},
		Username: user.Username,
		Role:     user.Role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtManner.secretKey))
}

func (jwtManner JWTManager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return []byte(jwtManner.secretKey), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (jwtManner JWTManager) IsCorrectPassword(password string, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}
