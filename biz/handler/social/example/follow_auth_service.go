package example

import (
	"VideoWeb/biz/dal/mysql"
	"VideoWeb/biz/dal/redis"
	format "VideoWeb/biz/handler/common_response_format"
	"context"
	"fmt"
	"net/http"
	"time"

	example0 "VideoWeb/biz/model/common/example"
	example "VideoWeb/biz/model/social/example"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/jwt"
)

// FollowAction .
// @router /api/follows [POST]
func FollowAction(ctx context.Context, c *app.RequestContext) {
	var req example.FollowRequest
	if err := c.BindAndValidate(&req); err != nil {
		format.Fail(c, http.StatusBadRequest, example0.ErrorCode_REQUEST_ERROR, err.Error())
		return
	}

	if req.UserID == 0 {
		format.Fail(c, http.StatusBadRequest, example0.ErrorCode_PARAM_ERROR, "目标用户ID不能为空")
		return
	}

	claims := jwt.ExtractClaims(ctx, c)
	userID, ok := claims["ID"].(float64)
	if !ok {
		format.Fail(c, http.StatusUnauthorized, example0.ErrorCode_USER_NOT_LOGIN, "用户未登录")
		return
	}

	currentUserID := int64(userID)

	if currentUserID == req.UserID {
		format.Fail(c, http.StatusBadRequest, example0.ErrorCode_PARAM_ERROR, "不能关注自己")
		return
	}

	db := mysql.GetDB()

	var targetUser mysql.Users
	if err := db.First(&targetUser, req.UserID).Error; err != nil {
		format.Fail(c, http.StatusNotFound, example0.ErrorCode_USER_NOT_EXIST, "目标用户不存在")
		return
	}

	var existingFollow mysql.Follows
	result := db.Where("follower_id = ? AND following_id = ?", currentUserID, req.UserID).
		Unscoped().
		First(&existingFollow)

	if req.Status {
		if result.Error == nil {
			if existingFollow.DeletedAt != nil {
				db.Model(&existingFollow).Update("deleted_at", nil)
			}
			format.Success(c, "follow", map[string]interface{}{
				"message":      "关注成功",
				"is_following": true,
			})
			return
		}

		newFollow := mysql.Follows{
			FollowerID:  currentUserID,
			FollowingID: req.UserID,
		}
		if err := db.Create(&newFollow).Error; err != nil {
			format.Fail(c, http.StatusInternalServerError, example0.ErrorCode_PROGRESS_ERROR, "关注失败")
			return
		}

		redis.Delete(fmt.Sprintf("follow:list:%d:*", currentUserID))
		redis.Delete(fmt.Sprintf("follower:list:%d:*", req.UserID))

		format.Success(c, "follow", map[string]interface{}{
			"message":      "关注成功",
			"is_following": true,
		})
	} else {
		if result.Error != nil {
			format.Fail(c, http.StatusBadRequest, example0.ErrorCode_PARAM_ERROR, "未关注该用户")
			return
		}

		if existingFollow.DeletedAt == nil {
			if err := db.Delete(&existingFollow).Error; err != nil {
				format.Fail(c, http.StatusInternalServerError, example0.ErrorCode_PROGRESS_ERROR, "取消关注失败")
				return
			}
		}

		redis.Delete(fmt.Sprintf("follow:list:%d:*", currentUserID))
		redis.Delete(fmt.Sprintf("follower:list:%d:*", req.UserID))

		format.Success(c, "unfollow", map[string]interface{}{
			"message":      "取消关注成功",
			"is_following": false,
		})
	}
}

// GetFriendList .
// @router /api/users/friends [GET]
func GetFriendList(ctx context.Context, c *app.RequestContext) {
	var req example.FriendListRequest
	if err := c.BindAndValidate(&req); err != nil {
		format.Fail(c, http.StatusBadRequest, example0.ErrorCode_REQUEST_ERROR, err.Error())
		return
	}

	claims := jwt.ExtractClaims(ctx, c)
	userID, ok := claims["ID"].(float64)
	if !ok {
		format.Fail(c, http.StatusUnauthorized, example0.ErrorCode_USER_NOT_LOGIN, "用户未登录")
		return
	}

	currentUserID := int64(userID)

	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	cacheKey := fmt.Sprintf("friend:list:%d:%d:%d", currentUserID, page, pageSize)

	type FriendListResult struct {
		Items []map[string]interface{}
		Total int64
	}
	var cachedResult FriendListResult
	err := redis.GetJSON(cacheKey, &cachedResult)
	if err == nil {
		format.Success(c, "friendList", map[string]interface{}{
			"items": cachedResult.Items,
			"total": cachedResult.Total,
		})
		return
	}
	if err.Error() != "redis not connected" {
		isNull, nullErr := redis.IsNullCache(cacheKey)
		if nullErr == nil && isNull {
			format.Success(c, "friendList", map[string]interface{}{
				"items": []map[string]interface{}{},
				"total": 0,
			})
			return
		}
	}

	db := mysql.GetDB()

	var friendIDs []int64
	db.Table("follows f1").
		Select("f1.following_id").
		Joins("INNER JOIN follows f2 ON f1.following_id = f2.follower_id AND f2.following_id = f1.follower_id").
		Where("f1.follower_id = ? AND f1.deleted_at IS NULL AND f2.deleted_at IS NULL", currentUserID).
		Pluck("f1.following_id", &friendIDs)

	total := int64(len(friendIDs))

	if len(friendIDs) == 0 {
		redis.SetNullCache(cacheKey, 10*time.Minute)
		format.Success(c, "friendList", map[string]interface{}{
			"items": []interface{}{},
			"total": 0,
		})
		return
	}

	offset := (page - 1) * pageSize
	end := offset + pageSize
	if end > len(friendIDs) {
		end = len(friendIDs)
	}
	if offset >= len(friendIDs) {
		format.Success(c, "friendList", map[string]interface{}{
           	"items": []interface{}{},
           	"total": total,
        })
		return
    }
	pagedFriendIDs := friendIDs[offset:end]

	var users []mysql.Users
	db.Table("users").
		Select("id, username, avatar_url, nickname, created_at").
		Where("id IN (?)", pagedFriendIDs).
		Find(&users)

	items := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		items = append(items, map[string]interface{}{
			"id":         u.ID,
			"username":   u.Username,
			"nickname":   u.Nickname,
			"avatar_url": u.Avatarurl,
		})
    }

    result := FriendListResult{
        Items: items,
        Total: total,
    }
    if len(items) == 0 {
		redis.SetNullCache(cacheKey, 10*time.Minute)
	} else {
		redis.SetJSON(cacheKey, result, 10*time.Minute)
	}

    format.Success(c, "friendList", map[string]interface{}{
		"items": items,
		"total": total,
	})
}
