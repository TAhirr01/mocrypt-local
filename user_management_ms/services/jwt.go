package services

import (
	"errors"
	"time"
	"user_management_ms/domain"
	"user_management_ms/dtos/response"

	"github.com/golang-jwt/jwt/v5"
)

type IJWTService interface {
	ParseJWT(tokenStr string) (*jwt.Token, error)
	GetClaims(token *jwt.Token) (jwt.MapClaims, error)
	GenerateToken(userID uint, duration time.Duration) (string, error)
	GenerateTokens(user *domain.User) (*response.Tokens, error)
}
type JWTService struct {
	Secret     []byte
	Issuer     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewJWTService(secret []byte, issuer string, accessTtl time.Duration, refreshTtl time.Duration) *JWTService {
	return &JWTService{
		Secret:     secret,
		Issuer:     issuer,
		AccessTTL:  accessTtl,
		RefreshTTL: refreshTtl,
	}
}

func (j *JWTService) ParseJWT(tokenStr string) (*jwt.Token, error) {
	if len(j.Secret) == 0 {
		return nil, errors.New("JWT secret is not configured")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.Secret, nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token, nil
}

func (j *JWTService) GetClaims(token *jwt.Token) (jwt.MapClaims, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["sub"] == nil {
		return nil, errors.New("No claims")
	}
	return claims, nil
}

func (j *JWTService) GenerateToken(userID uint, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iss": j.Issuer,
		"exp": time.Now().Add(duration).Unix(),
	})

	return token.SignedString(j.Secret)
}

func (j *JWTService) GenerateTokens(user *domain.User) (*response.Tokens, error) {
	accessToken, err := j.GenerateToken(user.Id, j.AccessTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, err := j.GenerateToken(user.Id, j.RefreshTTL)
	if err != nil {
		return nil, err
	}
	return &response.Tokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
