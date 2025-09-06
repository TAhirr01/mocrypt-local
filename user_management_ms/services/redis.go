package services

import (
	"context"
	"fmt"
	"time"
	"user_management_ms/config"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type IRedisService interface {
	SetRefreshToken(userId uint, refreshToken string) error
	GetRefreshToken(userId uint) (string, error)
	DelRefreshToken(userId uint)
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
