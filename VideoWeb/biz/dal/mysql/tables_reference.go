package mysql

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Users struct {
	gorm.Model
	Username  string `gorm:"column:username"`
	Password  string `gorm:"column:password"`
	Nickname  string `gorm:"column:nickname"`
	Avatarurl string `gorm:"column:avatar_url"`
}

type Videos struct {
	gorm.Model
	UserID      int64  `gorm:"column:user_id;index" json:"user_id"`
	Title       string `gorm:"column:title;size:100" json:"title"`
	Description string `gorm:"column:description;size:500" json:"description"`
	VideoURL    string `gorm:"column:video_url;size:500" json:"video_url"`
	Views       int64  `gorm:"column:views;default:0" json:"views"`
	Likes       int64  `gorm:"column:likes;default:0" json:"likes"`
	Comments    int64  `gorm:"column:comments;default:0" json:"comments"`
}

type Comments struct {
	gorm.Model
	UserID   int64  `gorm:"column:user_id;index" json:"user_id"`
	VideoID  int64  `gorm:"column:video_id;index" json:"video_id"`
	ParentID int64  `gorm:"column:parent_id;default:0;index" json:"parent_id"`
	Content  string `gorm:"column:content;size:500" json:"content"`
	Likes    int64  `gorm:"column:likes;default:0" json:"likes"`
}

type Likes struct {
	gorm.Model
	UserID   int64 `gorm:"column:user_id;index" json:"user_id"`
	TargetID int64 `gorm:"column:target_id;index" json:"target_id"`
	Type     int   `gorm:"column:type" json:"type"`
}

type Follows struct {
	ID          uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;index" json:"deleted_at"`
	FollowerID  int64      `gorm:"column:follower_id;uniqueIndex:uk_follow" json:"follower_id"`
	FollowingID int64      `gorm:"column:following_id;uniqueIndex:uk_follow" json:"following_id"`
}
