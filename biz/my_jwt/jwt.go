package my_jwt

import (
	"VideoWeb/biz/dal/mysql"
	"VideoWeb/biz/dal/redis"
	format "VideoWeb/biz/handler/common_response_format"
	example0 "VideoWeb/biz/model/common/example"
	"VideoWeb/biz/model/user/example"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hertz-contrib/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	identityKey    = "ID"
	AuthMiddleware *jwt.HertzJWTMiddleware
	jwtSecret      = []byte("this-is-a-secret")
)

const (
	accessTokenTTL          = 15 * time.Minute
	defaultRefreshTokenTTL  = 7 * 24 * time.Hour
	rememberRefreshTokenTTL = 30 * 24 * time.Hour
	tokenTypeAccess         = "access"
	tokenTypeRefresh        = "refresh"
)

func parseUserID(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return id, true
	default:
		return 0, false
	}
}

func refreshTokenKey(userID int64, jti string) string {
	return fmt.Sprintf("refresh:token:%d:%s", userID, jti)
}

func refreshTTL(remember bool) time.Duration {
	if remember {
		return rememberRefreshTokenTTL
	}
	return defaultRefreshTokenTTL
}

func GenerateAccessToken(userID int64) (string, time.Time, error) {
	expire := time.Now().Add(accessTokenTTL)
	claims := jwtv4.MapClaims{
		identityKey:  userID,
		"permission": "all",
		"typ":        tokenTypeAccess,
		"exp":        expire.Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expire, nil
}

func GenerateRefreshToken(userID int64, remember bool) (string, string, time.Time, error) {
	expire := time.Now().Add(refreshTTL(remember))
	jti := uuid.NewString()
	claims := jwtv4.MapClaims{
		identityKey: userID,
		"typ":       tokenTypeRefresh,
		"jti":       jti,
		"exp":       expire.Unix(),
		"iat":       time.Now().Unix(),
	}
	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", "", time.Time{}, err
	}
	return tokenString, jti, expire, nil
}

func StoreRefreshToken(userID int64, jti string, expire time.Time) error {
	ttl := time.Until(expire)
	if ttl <= 0 {
		return errors.New("refresh token expired")
	}
	return redis.SetWithExpiration(refreshTokenKey(userID, jti), "1", ttl)
}

func parseRefreshToken(tokenString string) (int64, string, time.Time, error) {
	token, err := jwtv4.Parse(tokenString, func(token *jwtv4.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv4.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, "", time.Time{}, errors.New("invalid refresh token")
	}
	claims, ok := token.Claims.(jwtv4.MapClaims)
	if !ok {
		return 0, "", time.Time{}, errors.New("invalid token claims")
	}
	tokenType, ok := claims["typ"].(string)
	if !ok || tokenType != tokenTypeRefresh {
		return 0, "", time.Time{}, errors.New("invalid token type")
	}
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return 0, "", time.Time{}, errors.New("invalid token jti")
	}
	userID, ok := parseUserID(claims[identityKey])
	if !ok {
		return 0, "", time.Time{}, errors.New("invalid token user")
	}
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return 0, "", time.Time{}, errors.New("invalid token expire")
	}
	expire := time.Unix(int64(expFloat), 0)
	return userID, jti, expire, nil
}

func RotateTokenPair(refreshToken string, remember bool) (string, time.Time, string, time.Time, int64, error) {
	userID, jti, _, err := parseRefreshToken(refreshToken)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, 0, err
	}
	exists, err := redis.Exists(refreshTokenKey(userID, jti))
	if err != nil {
		return "", time.Time{}, "", time.Time{}, 0, err
	}
	if !exists {
		return "", time.Time{}, "", time.Time{}, 0, errors.New("refresh token revoked")
	}
	if err = redis.Delete(refreshTokenKey(userID, jti)); err != nil {
		return "", time.Time{}, "", time.Time{}, 0, err
	}
	newRefreshToken, newJTI, refreshExpire, err := GenerateRefreshToken(userID, remember)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, 0, err
	}
	if err = StoreRefreshToken(userID, newJTI, refreshExpire); err != nil {
		return "", time.Time{}, "", time.Time{}, 0, err
	}
	accessToken, accessExpire, err := GenerateAccessToken(userID)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, 0, err
	}
	return accessToken, accessExpire, newRefreshToken, refreshExpire, userID, nil
}

func RevokeRefreshToken(refreshToken string) error {
	userID, jti, _, err := parseRefreshToken(refreshToken)
	if err != nil {
		return err
	}
	return redis.Delete(refreshTokenKey(userID, jti))
}

func InitJWT() error {
	var err error

	AuthMiddleware, err = jwt.New(&jwt.HertzJWTMiddleware{
		Realm:         "test zone",
		Key:           jwtSecret,
		Timeout:       accessTokenTTL,
		MaxRefresh:    accessTokenTTL,
		IdentityKey:   identityKey,
		TokenLookup:   "header: Authorization, query: token",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,

		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*mysql.Users); ok {
				return jwt.MapClaims{
					identityKey:  v.ID,
					"permission": "all",
					"typ":        tokenTypeAccess,
				}
			}
			return jwt.MapClaims{}
		},

		LoginResponse: func(ctx context.Context, c *app.RequestContext, code int, message string, expire time.Time) {
			userIDRaw, _ := c.Get(identityKey)
			userID, ok := parseUserID(userIDRaw)
			if !ok {
				format.Fail(c, 401, example0.ErrorCode_USER_NOT_LOGIN, "invalid login user")
				return
			}
			remember, _ := c.Get("remember")
			rememberLogin, _ := remember.(bool)
			refreshToken, refreshJTI, refreshExpire, err := GenerateRefreshToken(userID, rememberLogin)
			if err != nil {
				format.Fail(c, 500, example0.ErrorCode_PROGRESS_ERROR, "generate refresh token failed")
				return
			}
			if err = StoreRefreshToken(userID, refreshJTI, refreshExpire); err != nil {
				format.Fail(c, 500, example0.ErrorCode_PROGRESS_ERROR, "store refresh token failed")
				return
			}
			db := mysql.GetDB()
			userinfo := mysql.Users{}
			db.Table("users").
				Select("id, username, avatar_url, created_at, updated_at, nickname, deleted_at").
				Where("id = ?", userIDRaw).
				First(&userinfo)
			format.Success(c, "login", map[string]interface{}{
				"userinfo":       userinfo,
				"expire":         expire.Format("2006-01-02 15:04:05"),
				"access_token":   message,
				"access_expire":  expire.Format("2006-01-02 15:04:05"),
				"refresh_token":  refreshToken,
				"refresh_expire": refreshExpire.Format("2006-01-02 15:04:05"),
			})
		},

		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			var request example.UserLoginRequest
			if err := c.BindForm(&request); err != nil {
				return nil, err
			}

			if len(request.Username) == 0 {
				err = errors.New("username is missing")
				return nil, err
			}

			if len(request.Password) < 8 {
				err = errors.New("password is too short")
				return nil, err
			}
			db := mysql.GetDB()
			var userData mysql.Users
			db.Table("users").Where("username = ?", request.Username).First(&userData)
			if userData.Username == "" {
				err = errors.New("username is not exist")
				return nil, err
			}
			// 对比密码
			if err := bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(request.Password)); err != nil {
				err = errors.New("password is wrong")
				return nil, err
			}
			c.Set(identityKey, userData.ID)
			c.Set("remember", request.Remember)
			return &userData, nil
		},

		IdentityHandler: func(ctx context.Context, c *app.RequestContext) interface{} {
			claims := jwt.ExtractClaims(ctx, c)
			if permission, ok := claims["permission"].(string); ok {
				if tokenType, ok := claims["typ"].(string); ok && tokenType == tokenTypeAccess {
					return permission
				}
			}
			return ""
		},

		Authorizator: func(data interface{}, ctx context.Context, c *app.RequestContext) bool {
			permission, ok := data.(string)
			if !ok {
				return false
			}
			return permission == "all"
		},

		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			format.Fail(c, code, example0.ErrorCode_PARAM_ERROR, message)
		},
	})
	if err != nil {
		return err
	}
	return nil
}
