package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"telegram-notification/internal/config"
)

func TestHashPasswordStable(t *testing.T) {
	t.Parallel()
	a := HashPassword("abc123")
	b := HashPassword("abc123")
	if a != b {
		t.Fatalf("expected stable hash")
	}
}

func TestParseToken(t *testing.T) {
	t.Parallel()
	svc := NewAuthService(nil, config.AuthConfig{JWTSecret: "unit-secret", AccessTokenTTLMin: 60})
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid":   int64(7),
		"perms": []string{"bot.manage", "rule.manage"},
		"exp":   time.Now().Add(time.Minute).Unix(),
	})
	raw, err := token.SignedString([]byte("unit-secret"))
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}
	uid, perms, err := svc.ParseToken(raw)
	if err != nil {
		t.Fatalf("parse token failed: %v", err)
	}
	if uid != 7 {
		t.Fatalf("unexpected uid: %d", uid)
	}
	if !perms["bot.manage"] || !perms["rule.manage"] {
		t.Fatalf("permissions should be present")
	}
}
