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
	StoreSessionRedis(userID uint, sessionData *webauthn.SessionData) error
	GetSessionRedis(userID uint) (*webauthn.SessionData, error)
	DeleteSessionRedis(userID uint) error
	StoreRegistrationSessionRedis(sessionId string, sessionData *webauthn.SessionData) error
	GetRegistrationSessionRedis(sessionId string) (*webauthn.SessionData, error)
	DeleteRegistrationSessionRedis(sessionId string) error
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

func (s *RedisService) StoreSessionRedis(userID uint, sessionData *webauthn.SessionData) error {
	data, _ := json.Marshal(sessionData)
	return s.rdb.Set(ctx, fmt.Sprintf("webauthn:%d", userID), data, 5*time.Minute).Err()
}

// get session
func (s *RedisService) GetSessionRedis(userID uint) (*webauthn.SessionData, error) {
	val, err := s.rdb.Get(ctx, fmt.Sprintf("webauthn:%d", userID)).Result()
	if err != nil {
		return nil, err
	}

	var sessionData *webauthn.SessionData
	if err := json.Unmarshal([]byte(val), &sessionData); err != nil {
		return nil, err
	}
	return sessionData, nil
}

// delete session
func (s *RedisService) DeleteSessionRedis(userID uint) error {
	return s.rdb.Del(ctx, fmt.Sprintf("webauthn:%d", userID)).Err()
}

func (s *RedisService) StoreRegistrationSessionRedis(sessionId string, sessionData *webauthn.SessionData) error {
	data, _ := json.Marshal(sessionData)
	return s.rdb.Set(ctx, fmt.Sprintf("webauthn:%s", sessionId), data, 5*time.Minute).Err()
}
func (s *RedisService) GetRegistrationSessionRedis(sessionId string) (*webauthn.SessionData, error) {
	val, err := s.rdb.Get(ctx, fmt.Sprintf("webauthn:%s", sessionId)).Result()
	if err != nil {
		return nil, err
	}

	var sessionData *webauthn.SessionData
	if err := json.Unmarshal([]byte(val), &sessionData); err != nil {
		return nil, err
	}
	return sessionData, nil
}
func (s *RedisService) DeleteRegistrationSessionRedis(sessionId string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("webauthn:%d", sessionId)).Err()
}
