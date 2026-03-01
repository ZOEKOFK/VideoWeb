package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

// InitRedis 初始化Redis连接
func InitRedis() *redis.Client {
	Client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		fmt.Println("Redis连接失败: " + err.Error() + "，将不使用缓存功能")
	} else {
		fmt.Println("Redis连接成功(●'◡'●)")
	}

	return Client
}

func GetClient() *redis.Client {
	return Client
}

// Set 设置键值对
func Set(key string, value interface{}) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return Client.Set(Ctx, key, value, 0).Err()
}

// Get 获取键值
func Get(key string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("redis client not initialized")
	}
	return Client.Get(Ctx, key).Result()
}

// Delete 删除键
func Delete(key string) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return Client.Del(Ctx, key).Err()
}

// DeleteByPattern 根据模式删除键
func DeleteByPattern(pattern string) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	keys, err := Client.Keys(Ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return Client.Del(Ctx, keys...).Err()
	}
	return nil
}

// Exists 检查键是否存在
func Exists(key string) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("redis client not initialized")
	}
	result, err := Client.Exists(Ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// SetWithExpiration 设置带过期时间的键值对
func SetWithExpiration(key string, value interface{}, expiration time.Duration) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return Client.Set(Ctx, key, value, expiration).Err()
}

// Incr 自增操作
func Incr(key string) (int64, error) {
	if Client == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}
	return Client.Incr(Ctx, key).Result()
}

// Decr 自减操作
func Decr(key string) (int64, error) {
	if Client == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}
	return Client.Decr(Ctx, key).Result()
}

// HashSet 设置哈希表字段
func HashSet(key string, fieldValues ...interface{}) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return Client.HSet(Ctx, key, fieldValues...).Err()
}

// HashGet 获取哈希表字段
func HashGet(key, field string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("redis client not initialized")
	}
	return Client.HGet(Ctx, key, field).Result()
}

// HashDelete 删除哈希表字段
func HashDelete(key string, fields ...string) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return Client.HDel(Ctx, key, fields...).Err()
}

// SetJSON 将结构体序列化为JSON并存储
func SetJSON(key string, value interface{}, expiration time.Duration) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}
	return Client.Set(Ctx, key, jsonData, expiration).Err()
}

// GetJSON 从Redis获取JSON并反序列化为结构体
func GetJSON(key string, dest interface{}) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	data, err := Client.Get(Ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}
