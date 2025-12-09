package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gate-v2/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to connect to redis: %w", err)
	}

	return &RedisClient{Client: client}, nil
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// Cache operations
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}
	return r.Client.Set(ctx, key, data, expiration).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("marshal value: %w", err)
	}
	return r.Client.SetNX(ctx, key, data, expiration).Result()
}

// Hash operations
func (r *RedisClient) HSet(ctx context.Context, key string, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}
	return r.Client.HSet(ctx, key, field, data).Err()
}

func (r *RedisClient) HGet(ctx context.Context, key string, field string, dest interface{}) error {
	data, err := r.Client.HGet(ctx, key, field).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return r.Client.HDel(ctx, key, fields...).Err()
}

// List operations
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.LPush(ctx, key, values...).Err()
}

func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.RPush(ctx, key, values...).Err()
}

func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.LRange(ctx, key, start, stop).Result()
}

func (r *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return r.Client.LLen(ctx, key).Result()
}

// Set operations
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SAdd(ctx, key, members...).Err()
}

func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.Client.SMembers(ctx, key).Result()
}

func (r *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.Client.SIsMember(ctx, key, member).Result()
}

func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SRem(ctx, key, members...).Err()
}

// Counter operations
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

func (r *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.Client.IncrBy(ctx, key, value).Result()
}

func (r *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return r.Client.Decr(ctx, key).Result()
}

func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}

// Pub/Sub
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	return r.Client.Publish(ctx, channel, data).Err()
}

func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.Client.Subscribe(ctx, channels...)
}

// Cache key builders
const (
	CacheKeyUserPrefix        = "user:"
	CacheKeyAdminPrefix       = "admin:"
	CacheKeySessionPrefix     = "session:"
	CacheKeyProductPrefix     = "product:"
	CacheKeySKUPrefix         = "sku:"
	CacheKeyRegionPrefix      = "region:"
	CacheKeyPromoPrefix       = "promo:"
	CacheKeyRateLimitPrefix   = "ratelimit:"
	CacheKeyValidationPrefix  = "validation:"
	CacheKeyMFAPrefix         = "mfa:"
)

func (r *RedisClient) UserCacheKey(userID string) string {
	return CacheKeyUserPrefix + userID
}

func (r *RedisClient) AdminCacheKey(adminID string) string {
	return CacheKeyAdminPrefix + adminID
}

func (r *RedisClient) SessionCacheKey(sessionID string) string {
	return CacheKeySessionPrefix + sessionID
}

func (r *RedisClient) ProductCacheKey(productCode, region string) string {
	return fmt.Sprintf("%s%s:%s", CacheKeyProductPrefix, productCode, region)
}

func (r *RedisClient) SKUCacheKey(skuCode, region string) string {
	return fmt.Sprintf("%s%s:%s", CacheKeySKUPrefix, skuCode, region)
}

func (r *RedisClient) RegionCacheKey(code string) string {
	return CacheKeyRegionPrefix + code
}

func (r *RedisClient) PromoCacheKey(promoCode string) string {
	return CacheKeyPromoPrefix + promoCode
}

func (r *RedisClient) RateLimitKey(identifier, endpoint string) string {
	return fmt.Sprintf("%s%s:%s", CacheKeyRateLimitPrefix, identifier, endpoint)
}

func (r *RedisClient) ValidationTokenKey(token string) string {
	return CacheKeyValidationPrefix + token
}

func (r *RedisClient) MFATokenKey(token string) string {
	return CacheKeyMFAPrefix + token
}

// Cache invalidation helpers
func (r *RedisClient) InvalidateUserCache(ctx context.Context, userID string) error {
	pattern := r.UserCacheKey(userID) + "*"
	return r.deleteByPattern(ctx, pattern)
}

func (r *RedisClient) InvalidateProductCache(ctx context.Context, productCode string) error {
	pattern := CacheKeyProductPrefix + productCode + "*"
	return r.deleteByPattern(ctx, pattern)
}

func (r *RedisClient) InvalidateSKUCache(ctx context.Context, skuCode string) error {
	pattern := CacheKeySKUPrefix + skuCode + "*"
	return r.deleteByPattern(ctx, pattern)
}

func (r *RedisClient) deleteByPattern(ctx context.Context, pattern string) error {
	iter := r.Client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.Client.Del(ctx, keys...).Err()
	}
	return nil
}

