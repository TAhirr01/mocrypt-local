package services

import (
	"errors"
	"fmt"
	"user_management_ms/dtos/response"
	"user_management_ms/repository/query_repository"

	"github.com/hashicorp/go-uuid"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

type IQRLoginService interface {
	RequestLoginQr() ([]byte, string, error)
	ApproveLoginQr(userId uint, sessionId string) error
	CheckLoginQr(sessionId string) (*response.QrLoginResponse, error)
}

type QRLoginService struct {
	redis IRedisService
	db    *gorm.DB
	jwt   IJWTService
	query query_repository.IUserQueryRepository
}

func NewQRLoginService(redis IRedisService, jwt IJWTService, query query_repository.IUserQueryRepository, db *gorm.DB) IQRLoginService {
	return &QRLoginService{redis: redis, jwt: jwt, query: query, db: db}
}

func (u *QRLoginService) RequestLoginQr() ([]byte, string, error) {
	sessionId, _ := uuid.GenerateUUID()
	err := u.redis.StoreLoginSessionRedis(sessionId)
	if err != nil {
		return nil, "", err
	}
	url := fmt.Sprintf("https://mocadomain.com/qr-login?sessionId=%s", sessionId)
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return nil, "", err
	}

	return png, sessionId, nil
}

func (u *QRLoginService) ApproveLoginQr(userId uint, sessionId string) error {
	session, err := u.redis.GetLoginSessionRedis(sessionId)
	if err != nil {
		return errors.New("session not found or redis problem")
	}
	session.UserId = userId
	session.Status = "APPROVED"

	if err := u.redis.UpdateLoginSessionRedis(sessionId, session); err != nil {
		return err
	}
	return nil
}

func (u *QRLoginService) CheckLoginQr(sessionId string) (*response.QrLoginResponse, error) {
	session, err := u.redis.GetLoginSessionRedis(sessionId)
	if err != nil {
		return &response.QrLoginResponse{Status: response.StatusExpired}, nil
	}

	switch session.Status {
	case "PENDING":
		return &response.QrLoginResponse{Status: response.StatusPending}, nil
	case "APPROVED":
		user, err := u.query.GetByID(u.db, session.UserId)
		if err != nil {
			return nil, err
		}
		tokens, err := u.jwt.GenerateTokens(user)
		if err != nil {
			return nil, err
		}

		// Once consumed, delete to avoid reuse
		_ = u.redis.DeleteLoginSessionRedis(sessionId)

		return &response.QrLoginResponse{
			Status: response.StatusApproved,
			Tokens: tokens,
		}, nil

	default:
		return &response.QrLoginResponse{Status: response.StatusExpired}, nil
	}
}
