// Package jwt 提供 JWT（JSON Web Token）令牌的生成与解析功能。
//
// 使用 HMAC-SHA256 算法签名，Token 中包含：
//   - 用户ID、UUID、用户名、角色列表
//   - 标准 JWT 声明（签发者、签发时间、过期时间、生效时间）
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"casbin-demo/internal/config"
)

// Claims 自定义 JWT 声明，在标准声明基础上扩展业务字段。
type Claims struct {
	UserID   uint     `json:"user_id"`  // 用户ID
	UUID     string   `json:"uuid"`     // 用户UUID
	Username string   `json:"username"` // 用户名（Casbin 策略 sub 使用此字段）
	Roles    []string `json:"roles"`    // 用户角色列表
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT 令牌。
//
// 参数：
//   - userID: 用户ID
//   - uuid: 用户UUID
//   - username: 用户名
//   - roles: 用户角色名称列表
//
// 令牌有效期由 config.JWT.ExpireHours 控制（默认 24 小时）。
func GenerateToken(userID uint, uuid, username string, roles []string) (string, error) {
	cfg := config.GlobalConfig
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		UUID:     uuid,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.JWT.ExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    cfg.JWT.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ParseToken 解析并验证 JWT 令牌。
//
// 返回解析后的 Claims，若令牌无效、过期或签名错误则返回 error。
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.GlobalConfig
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
