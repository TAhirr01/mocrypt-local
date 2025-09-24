package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"user_management_ms/config"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type IRedisService interface {
	SetRefreshToken(userId uint, refreshToken string) error
	GetRefreshToken(userId uint) (string, error)
	DelRefreshToken(userId uint)
	StoreSessionRedis(sessionId string, sessionData *webauthn.SessionData) error
	GetSessionRedis(sessionId string) (*webauthn.SessionData, error)
	DeleteSessionRedis(sessionId string) error
	StoreRegistrationSessionRedis(userID uint, sessionData *webauthn.SessionData) error
	GetRegistrationSessionRedis(userID uint) (*webauthn.SessionData, error)
	DeleteRegistrationSessionRedis(userID uint) error
	StoreLoginSessionRedis(sessionId string) error
	GetLoginSessionRedis(sessionId string) (*RedisSession, error)
	DeleteLoginSessionRedis(sessionId string) error
	UpdateLoginSessionRedis(sessionId string, session *RedisSession) error
}

type RedisSession struct {
	SessionId string `json:"sessionId"`
	Status    string `json:"status"`
	UserId    uint   `json:"userId"`
}
type RedisService struct {
	rdb *redis.Client
}

func NewRedisService(rdb *redis.Client) *RedisService {
	return &RedisService{rdb: rdb}
}

func (s *RedisService) SetRefreshToken(userId uint, refreshToken string) error {
	return s.rdb.SetEx(ctx, fmt.Sprintf("refresh_%d", userId), refreshToken, time.Duration(config.Conf.Application.Security.TokenValidityInSecondsForRememberMe)*time.Second).Err()
}

func (s *RedisService) GetRefreshToken(userId uint) (string, error) {
	return s.rdb.Get(ctx, fmt.Sprintf("refresh_%d", userId)).Result()
}

func (s *RedisService) DelRefreshToken(userId uint) {
	s.rdb.Del(ctx, fmt.Sprintf("refresh_%d", userId))
}

func (s *RedisService) StoreSessionRedis(sessionId string, sessionData *webauthn.SessionData) error {
	data, _ := json.Marshal(sessionData)
	return s.rdb.Set(ctx, fmt.Sprintf("webauthn:%s", sessionId), data, 5*time.Minute).Err()
}

// get session
func (s *RedisService) GetSessionRedis(sessionId string) (*webauthn.SessionData, error) {
	val, err := s.rdb.Get(ctx, fmt.Sprintf("webauthn:%s", sessionId)).Result()
	if err != nil {
		return nil, err
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(val), &sessionData); err != nil {
		return nil, err
	}
	return &sessionData, nil
}

// delete session
func (s *RedisService) DeleteSessionRedis(sessionId string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("webauthn:%s", sessionId)).Err()
}

func (s *RedisService) StoreRegistrationSessionRedis(userId uint, sessionData *webauthn.SessionData) error {
	data, _ := json.Marshal(sessionData)
	return s.rdb.Set(ctx, fmt.Sprintf("webauthn:%d", userId), data, 5*time.Minute).Err()
}
func (s *RedisService) GetRegistrationSessionRedis(userId uint) (*webauthn.SessionData, error) {
	val, err := s.rdb.Get(ctx, fmt.Sprintf("webauthn:%d", userId)).Result()
	if err != nil {
		return nil, err
	}
	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(val), &sessionData); err != nil {
		return nil, err
	}
	return &sessionData, nil
}
func (s *RedisService) DeleteRegistrationSessionRedis(userId uint) error {
	return s.rdb.Del(ctx, fmt.Sprintf("webauthn:%d", userId)).Err()
}

func (s *RedisService) StoreLoginSessionRedis(sessionId string) error {
	redisSession := &RedisSession{
		Status: "PENDING",
		UserId: 0,
	}

	return s.rdb.Set(ctx, fmt.Sprintf("qrlogin:%s", sessionId), redisSession, 1*time.Minute).Err()
}

func (s *RedisService) GetLoginSessionRedis(sessionId string) (*RedisSession, error) {
	val, err := s.rdb.Get(ctx, fmt.Sprintf("qrlogin:%s", sessionId)).Result()
	if err != nil {
		return nil, err
	}
	var redisSession RedisSession

	if err := json.Unmarshal([]byte(val), &redisSession); err != nil {
		return nil, err
	}
	return &redisSession, err
}

func (s *RedisService) DeleteLoginSessionRedis(sessionId string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("qrlogin:%s", sessionId)).Err()
}

func (s *RedisService) UpdateLoginSessionRedis(sessionId string, session *RedisSession) error {
	s.rdb.Del(ctx, fmt.Sprintf("qrlogin:%s", sessionId))
	return s.rdb.Set(ctx, fmt.Sprintf("qrlogin:%s", sessionId), session, 1*time.Minute).Err()
}
