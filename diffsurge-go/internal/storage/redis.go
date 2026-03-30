package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/go-redis/redis/v8"
)

const (
	trafficQueueKey = "diffsurge:queue:traffic"
	cacheKeyPrefix  = "diffsurge:cache:"
	rateLimitPrefix = "diffsurge:ratelimit:"

	// Default TTLs for caching
	ProjectCacheTTL      = 5 * time.Minute
	EnvironmentCacheTTL  = 5 * time.Minute
	StatsCacheTTL        = 30 * time.Second
	UserOrgsCacheTTL     = 5 * time.Minute
	SubscriptionLimitTTL = 1 * time.Minute
)

// RedisStore handles Redis operations for caching and buffering
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new Redis store
func NewRedisStore(redisURL string) (*RedisStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{client: client}, nil
}

// Close closes the Redis connection
func (r *RedisStore) Close() error {
	return r.client.Close()
}

// Ping checks if Redis is reachable
func (r *RedisStore) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// ===== Traffic Queue Operations =====

// EnqueueTraffic pushes a traffic log to the queue (LPUSH)
func (r *RedisStore) EnqueueTraffic(ctx context.Context, log *models.TrafficLog) error {
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal traffic log: %w", err)
	}

	if err := r.client.LPush(ctx, trafficQueueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to push to queue: %w", err)
	}

	return nil
}

// DequeueTraffic pops a traffic log from the queue (BRPOP with timeout)
func (r *RedisStore) DequeueTraffic(ctx context.Context, timeout time.Duration) (*models.TrafficLog, error) {
	result, err := r.client.BRPop(ctx, timeout, trafficQueueKey).Result()
	if err == redis.Nil {
		// Timeout - no items in queue
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to pop from queue: %w", err)
	}

	// result[0] is the key, result[1] is the value
	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected result from BRPOP")
	}

	var log models.TrafficLog
	if err := json.Unmarshal([]byte(result[1]), &log); err != nil {
		return nil, fmt.Errorf("failed to unmarshal traffic log: %w", err)
	}

	return &log, nil
}

// GetQueueLength returns the number of items in the traffic queue
func (r *RedisStore) GetQueueLength(ctx context.Context) (int64, error) {
	return r.client.LLen(ctx, trafficQueueKey).Result()
}

// ===== Caching Operations =====

// SetCache stores a value in the cache with TTL
func (r *RedisStore) SetCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	cacheKey := cacheKeyPrefix + key
	if err := r.client.Set(ctx, cacheKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// GetCache retrieves a value from the cache
func (r *RedisStore) GetCache(ctx context.Context, key string, dest interface{}) error {
	cacheKey := cacheKeyPrefix + key
	data, err := r.client.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return ErrCacheNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to get cache: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// DeleteCache removes a value from the cache
func (r *RedisStore) DeleteCache(ctx context.Context, key string) error {
	cacheKey := cacheKeyPrefix + key
	return r.client.Del(ctx, cacheKey).Err()
}

// DeleteCachePattern removes all keys matching a pattern
func (r *RedisStore) DeleteCachePattern(ctx context.Context, pattern string) error {
	matchPattern := cacheKeyPrefix + pattern
	iter := r.client.Scan(ctx, 0, matchPattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("scan iteration error: %w", err)
	}

	return nil
}

// ===== Rate Limiting Operations =====

// IncrementRateLimit increments the rate limit counter for a key
// Uses sliding window approach with sorted sets
func (r *RedisStore) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	rateLimitKey := rateLimitPrefix + key
	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	pipe := r.client.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, rateLimitKey, "0", fmt.Sprintf("%d", windowStart))

	// Add current request
	pipe.ZAdd(ctx, rateLimitKey, &redis.Z{
		Score:  float64(now),
		Member: now,
	})

	// Count requests in window
	countCmd := pipe.ZCount(ctx, rateLimitKey, fmt.Sprintf("%d", windowStart), "+inf")

	// Set expiration to prevent memory leaks
	pipe.Expire(ctx, rateLimitKey, window+time.Minute)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	count, err := countCmd.Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get count: %w", err)
	}

	return count, nil
}

// CheckRateLimit checks if a key is within the rate limit
func (r *RedisStore) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, error) {
	count, err := r.IncrementRateLimit(ctx, key, window)
	if err != nil {
		return false, 0, err
	}

	return count <= limit, count, nil
}

// GetRateLimitCount returns the current count for a rate limit key
func (r *RedisStore) GetRateLimitCount(ctx context.Context, key string, window time.Duration) (int64, error) {
	rateLimitKey := rateLimitPrefix + key
	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	return r.client.ZCount(ctx, rateLimitKey, fmt.Sprintf("%d", windowStart), "+inf").Result()
}

// ResetRateLimit resets the rate limit counter for a key
func (r *RedisStore) ResetRateLimit(ctx context.Context, key string) error {
	rateLimitKey := rateLimitPrefix + key
	return r.client.Del(ctx, rateLimitKey).Err()
}

// ===== Pub/Sub Operations =====

// Publish publishes a message to a channel
func (r *RedisStore) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := r.client.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Subscribe subscribes to a channel and returns a subscription
func (r *RedisStore) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// ===== Helper Functions =====

// BuildProjectCacheKey builds a cache key for a project
func BuildProjectCacheKey(projectID string) string {
	return fmt.Sprintf("project:%s", projectID)
}

// BuildEnvironmentCacheKey builds a cache key for an environment
func BuildEnvironmentCacheKey(envID string) string {
	return fmt.Sprintf("env:%s", envID)
}

// BuildStatsCacheKey builds a cache key for traffic stats
func BuildStatsCacheKey(projectID string, period string) string {
	return fmt.Sprintf("traffic:stats:%s:%s", projectID, period)
}

// BuildUserOrgsCacheKey builds a cache key for user organizations
func BuildUserOrgsCacheKey(userID string) string {
	return fmt.Sprintf("user:%s:orgs", userID)
}

// BuildSubscriptionLimitsCacheKey builds a cache key for subscription limits
func BuildSubscriptionLimitsCacheKey(orgID string) string {
	return fmt.Sprintf("sub:%s:limits", orgID)
}

// ErrCacheNotFound is returned when a cache key is not found
var ErrCacheNotFound = fmt.Errorf("cache key not found")
