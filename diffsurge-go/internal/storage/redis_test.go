package storage

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*RedisStore, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	store := &RedisStore{
		client: client,
	}

	return store, mr
}

func TestRedisStore_EnqueueDequeueTraffic(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	log := &models.TrafficLog{
		ID:         uuid.New(),
		ProjectID:  uuid.New(),
		Method:     "GET",
		Path:       "/api/users",
		StatusCode: 200,
		Timestamp:  time.Now(),
	}

	// Enqueue
	err := store.EnqueueTraffic(ctx, log)
	require.NoError(t, err)

	// Check queue length
	length, err := store.GetQueueLength(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), length)

	// Dequeue
	retrieved, err := store.DequeueTraffic(ctx, 1*time.Second)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, log.ID, retrieved.ID)
	assert.Equal(t, log.Method, retrieved.Method)
	assert.Equal(t, log.Path, retrieved.Path)

	// Queue should be empty now
	length, err = store.GetQueueLength(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), length)
}

func TestRedisStore_DequeueTimeout(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	// Dequeue from empty queue should return nil after timeout
	start := time.Now()
	retrieved, err := store.DequeueTraffic(ctx, 100*time.Millisecond)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Nil(t, retrieved)
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
}

func TestRedisStore_MultipleTrafficLogs(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	logs := []*models.TrafficLog{
		{ID: uuid.New(), Method: "GET", Path: "/api/users"},
		{ID: uuid.New(), Method: "POST", Path: "/api/orders"},
		{ID: uuid.New(), Method: "DELETE", Path: "/api/products"},
	}

	// Enqueue all
	for _, log := range logs {
		err := store.EnqueueTraffic(ctx, log)
		require.NoError(t, err)
	}

	// Check queue length
	length, err := store.GetQueueLength(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), length)

	// Dequeue in FIFO order (first in, first out)
	for i := 0; i < len(logs); i++ {
		retrieved, err := store.DequeueTraffic(ctx, 1*time.Second)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, logs[i].ID, retrieved.ID)
	}
}

func TestRedisStore_SetGetCache(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	type TestData struct {
		Name  string
		Value int
	}

	data := TestData{Name: "test", Value: 42}
	key := "test:key"

	// Set cache
	err := store.SetCache(ctx, key, data, 1*time.Minute)
	require.NoError(t, err)

	// Get cache
	var retrieved TestData
	err = store.GetCache(ctx, key, &retrieved)
	require.NoError(t, err)
	assert.Equal(t, data.Name, retrieved.Name)
	assert.Equal(t, data.Value, retrieved.Value)
}

func TestRedisStore_CacheNotFound(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	var data interface{}
	err := store.GetCache(ctx, "nonexistent", &data)
	assert.ErrorIs(t, err, ErrCacheNotFound)
}

func TestRedisStore_CacheExpiration(t *testing.T) {
	store, mr := setupTestRedis(t)
	ctx := context.Background()

	data := map[string]string{"key": "value"}
	key := "expiring:key"

	// Set cache with short TTL
	err := store.SetCache(ctx, key, data, 100*time.Millisecond)
	require.NoError(t, err)

	// Should exist initially
	var retrieved map[string]string
	err = store.GetCache(ctx, key, &retrieved)
	require.NoError(t, err)

	// Fast-forward time in miniredis
	mr.FastForward(200 * time.Millisecond)

	// Should be expired now
	err = store.GetCache(ctx, key, &retrieved)
	assert.ErrorIs(t, err, ErrCacheNotFound)
}

func TestRedisStore_DeleteCache(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	data := "test data"
	key := "deletable:key"

	// Set cache
	err := store.SetCache(ctx, key, data, 1*time.Minute)
	require.NoError(t, err)

	// Verify it exists
	var retrieved string
	err = store.GetCache(ctx, key, &retrieved)
	require.NoError(t, err)

	// Delete
	err = store.DeleteCache(ctx, key)
	require.NoError(t, err)

	// Should not exist now
	err = store.GetCache(ctx, key, &retrieved)
	assert.ErrorIs(t, err, ErrCacheNotFound)
}

func TestRedisStore_DeleteCachePattern(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	// Set multiple keys with same prefix
	for i := 0; i < 5; i++ {
		key := "project:test:" + string(rune('a'+i))
		err := store.SetCache(ctx, key, i, 1*time.Minute)
		require.NoError(t, err)
	}

	// Set a key with different prefix
	err := store.SetCache(ctx, "other:key", "other", 1*time.Minute)
	require.NoError(t, err)

	// Delete all project:test:* keys
	err = store.DeleteCachePattern(ctx, "project:test:*")
	require.NoError(t, err)

	// Verify project keys are deleted
	var val int
	err = store.GetCache(ctx, "project:test:a", &val)
	assert.ErrorIs(t, err, ErrCacheNotFound)

	// Verify other key still exists
	var other string
	err = store.GetCache(ctx, "other:key", &other)
	require.NoError(t, err)
	assert.Equal(t, "other", other)
}

func TestRedisStore_RateLimit(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	key := "test:ratelimit"
	limit := int64(5)
	window := 1 * time.Second

	// Should allow first 5 requests
	for i := int64(0); i < limit; i++ {
		allowed, count, err := store.CheckRateLimit(ctx, key, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed, "request %d should be allowed", i+1)
		assert.Equal(t, i+1, count)
	}

	// 6th request should be denied
	allowed, count, err := store.CheckRateLimit(ctx, key, limit, window)
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, int64(6), count)
}

func TestRedisStore_RateLimitSlidingWindow(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	key := "test:sliding"
	limit := int64(3)
	window := 200 * time.Millisecond

	// Use 3 requests
	for i := 0; i < 3; i++ {
		allowed, _, err := store.CheckRateLimit(ctx, key, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// 4th should fail
	allowed, _, err := store.CheckRateLimit(ctx, key, limit, window)
	require.NoError(t, err)
	assert.False(t, allowed)

	// Wait for window to pass (use real sleep instead of FastForward)
	time.Sleep(250 * time.Millisecond)

	// Should allow new requests now (old entries cleaned up)
	allowed, count, err := store.CheckRateLimit(ctx, key, limit, window)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, int64(1), count)
}

func TestRedisStore_GetRateLimitCount(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	key := "test:count"
	window := 1 * time.Second

	// No requests yet
	count, err := store.GetRateLimitCount(ctx, key, window)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Add some requests
	for i := 0; i < 3; i++ {
		_, _, err := store.CheckRateLimit(ctx, key, 10, window)
		require.NoError(t, err)
	}

	// Check count
	count, err = store.GetRateLimitCount(ctx, key, window)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestRedisStore_ResetRateLimit(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	key := "test:reset"
	limit := int64(2)
	window := 1 * time.Second

	// Use up the limit
	for i := 0; i < 2; i++ {
		_, _, _ = store.CheckRateLimit(ctx, key, limit, window)
	}

	// Should be at limit
	allowed, _, err := store.CheckRateLimit(ctx, key, limit, window)
	require.NoError(t, err)
	assert.False(t, allowed)

	// Reset
	err = store.ResetRateLimit(ctx, key)
	require.NoError(t, err)

	// Should allow requests again
	allowed, count, err := store.CheckRateLimit(ctx, key, limit, window)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, int64(1), count)
}

func TestRedisStore_Ping(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	err := store.Ping(ctx)
	assert.NoError(t, err)
}

func TestRedisStore_Close(t *testing.T) {
	store, _ := setupTestRedis(t)

	err := store.Close()
	assert.NoError(t, err)

	// Operations after close should fail
	ctx := context.Background()
	err = store.Ping(ctx)
	assert.Error(t, err)
}

func TestBuildCacheKeys(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "project cache key",
			fn:       func() string { return BuildProjectCacheKey("proj-123") },
			expected: "project:proj-123",
		},
		{
			name:     "environment cache key",
			fn:       func() string { return BuildEnvironmentCacheKey("env-456") },
			expected: "env:env-456",
		},
		{
			name:     "stats cache key",
			fn:       func() string { return BuildStatsCacheKey("proj-123", "30d") },
			expected: "traffic:stats:proj-123:30d",
		},
		{
			name:     "user orgs cache key",
			fn:       func() string { return BuildUserOrgsCacheKey("user-789") },
			expected: "user:user-789:orgs",
		},
		{
			name:     "subscription limits cache key",
			fn:       func() string { return BuildSubscriptionLimitsCacheKey("org-abc") },
			expected: "sub:org-abc:limits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedisStore_PublishSubscribe(t *testing.T) {
	store, _ := setupTestRedis(t)
	ctx := context.Background()

	channel := "test:channel"
	message := map[string]string{"event": "test", "data": "hello"}

	// Subscribe
	pubsub := store.Subscribe(ctx, channel)
	defer pubsub.Close()

	// Wait for subscription to be ready
	_, err := pubsub.Receive(ctx)
	require.NoError(t, err)

	// Publish
	err = store.Publish(ctx, channel, message)
	require.NoError(t, err)

	// Receive message
	msg, err := pubsub.ReceiveMessage(ctx)
	require.NoError(t, err)
	assert.Equal(t, channel, msg.Channel)

	// Should contain the marshaled message
	assert.Contains(t, msg.Payload, "test")
	assert.Contains(t, msg.Payload, "hello")
}

// Benchmark tests
func BenchmarkRedisStore_EnqueueTraffic(b *testing.B) {
	mr := miniredis.NewMiniRedis()
	if err := mr.Start(); err != nil {
		b.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	store := &RedisStore{client: client}
	ctx := context.Background()

	log := &models.TrafficLog{
		ID:         uuid.New(),
		ProjectID:  uuid.New(),
		Method:     "GET",
		Path:       "/api/test",
		StatusCode: 200,
		Timestamp:  time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.EnqueueTraffic(ctx, log)
	}
}

func BenchmarkRedisStore_SetCache(b *testing.B) {
	mr := miniredis.NewMiniRedis()
	if err := mr.Start(); err != nil {
		b.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	store := &RedisStore{client: client}
	ctx := context.Background()

	data := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.SetCache(ctx, "bench:key", data, 1*time.Minute)
	}
}

func BenchmarkRedisStore_GetCache(b *testing.B) {
	mr := miniredis.NewMiniRedis()
	if err := mr.Start(); err != nil {
		b.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	store := &RedisStore{client: client}
	ctx := context.Background()

	data := map[string]interface{}{"key": "value"}
	_ = store.SetCache(ctx, "bench:key", data, 1*time.Minute)

	var retrieved map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.GetCache(ctx, "bench:key", &retrieved)
	}
}

func BenchmarkRedisStore_CheckRateLimit(b *testing.B) {
	mr := miniredis.NewMiniRedis()
	if err := mr.Start(); err != nil {
		b.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	store := &RedisStore{client: client}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = store.CheckRateLimit(ctx, "bench:limit", 1000, 1*time.Minute)
	}
}
