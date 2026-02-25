package my_jwt

import (
	"VideoWeb/biz/dal/mysql"
	format "VideoWeb/biz/handler/common_response_format"
	example0 "VideoWeb/biz/model/common/example"
	"VideoWeb/biz/model/user/example"
	"context"
	"errors"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	identityKey    = "ID"
	AuthMiddleware *jwt.HertzJWTMiddleware
)

func InitJWT() error {
	var err error

	AuthMiddleware, err = jwt.New(&jwt.HertzJWTMiddleware{
		Realm:         "test zone",
		Key:           []byte("this-is-a-secret"),
		Timeout:       time.Hour,
		MaxRefresh:    time.Hour,
		IdentityKey:   identityKey,
		TokenLookup:   "header: Authorization, query: token",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,

		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*mysql.Users); ok {
				return jwt.MapClaims{
					identityKey:  v.ID,
					"exp":        time.Now().Add(2 * time.Hour).Unix(),
					"permission": "all",
				}
			}
			return jwt.MapClaims{}
		},

		LoginResponse: func(ctx context.Context, c *app.RequestContext, code int, message string, expire time.Time) {
			claims := jwt.ExtractClaims(ctx, c)
			tokenUserID, _ := claims["ID"].(float64)
			userID := int64(tokenUserID)
			db := mysql.GetDB()
			userinfo := mysql.Users{}
			db.Table("users").
				Select("id, username, avatar_url, created_at, updated_at, nickname, deleted_at").
				Where("id = ?", userID).
				First(&userinfo)
			format.Success(c, "login", map[string]interface{}{
				"userinfo": userinfo,
				"token":    message,                              // Token字符串
				"expire":   expire.Format("2006-01-02 15:04:05"), // 过期时间（格式化）
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
			return &userData, nil
		},

		IdentityHandler: func(ctx context.Context, c *app.RequestContext) interface{} {
			claims := jwt.ExtractClaims(ctx, c)
			if permission, ok := claims["permission"].(string); ok {
				return permission
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
