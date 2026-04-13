package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/yclenove/telegram-relay/internal/config"
	"github.com/yclenove/telegram-relay/internal/repository/postgres"
)

// AuthService 负责后台账号认证与 JWT 签发。
type AuthService struct {
	store *postgres.Store
	cfg   config.AuthConfig
}

func NewAuthService(store *postgres.Store, cfg config.AuthConfig) *AuthService {
	return &AuthService{store: store, cfg: cfg}
}

func HashPassword(raw string) string {
	// 说明：当前为轻量实现，生产建议替换为 bcrypt/argon2。
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, []string, int64, error) {
	user, perms, err := s.store.FindUserWithPermissions(ctx, username)
	if err != nil {
		return "", nil, 0, err
	}
	if !user.IsEnabled {
		return "", nil, 0, errors.New("user disabled")
	}
	if user.PasswordHash != HashPassword(password) {
		return "", nil, 0, errors.New("invalid username or password")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid":   user.ID,
		"uname": user.Username,
		"perms": perms,
		"exp":   time.Now().Add(time.Duration(s.cfg.AccessTokenTTLMin) * time.Minute).Unix(),
		"iat":   time.Now().Unix(),
	})
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", nil, 0, err
	}
	// nil slice 序列化为 JSON null，前端读 permissions.length 会抛错，统一成空数组。
	if perms == nil {
		perms = []string{}
	}
	return signed, perms, user.ID, nil
}

func (s *AuthService) ParseToken(raw string) (int64, map[string]bool, error) {
	tok, err := jwt.Parse(raw, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !tok.Valid {
		return 0, nil, errors.New("invalid token")
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, nil, errors.New("invalid claims")
	}
	uidFloat, ok := claims["uid"].(float64)
	if !ok {
		return 0, nil, errors.New("invalid uid")
	}
	perms := map[string]bool{}
	if arr, ok := claims["perms"].([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				perms[s] = true
			}
		}
	}
	return int64(uidFloat), perms, nil
}
