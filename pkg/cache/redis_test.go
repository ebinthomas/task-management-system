package cache

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestRedis(t *testing.T) (*RedisCache, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}

	// Parse host and port from miniredis address
	addr := strings.Split(mr.Addr(), ":")
	if len(addr) != 2 {
		t.Fatalf("Invalid miniredis address format: %s", mr.Addr())
	}

	// Create Redis cache with the full address
	cache, err := NewRedisCache(mr.Addr(), "", 0)
	if err != nil {
		t.Fatalf("Failed to create Redis cache: %v", err)
	}

	return cache, mr
}

func TestRedisCache_Set(t *testing.T) {
	cache, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	testData := "test_value"

	// Test successful set
	err := cache.Set(ctx, "test_key", testData, 1*time.Hour)
	assert.NoError(t, err)

	// Verify value was set (note: Redis stores JSON-encoded values)
	val, err := mr.Get("test_key")
	assert.NoError(t, err)
	var result string
	err = json.Unmarshal([]byte(val), &result)
	assert.NoError(t, err)
	assert.Equal(t, testData, result)

	// Test set with invalid connection
	mr.Close() // Close Redis to simulate connection error
	err = cache.Set(ctx, "test_key2", testData, 1*time.Hour)
	assert.Error(t, err)
}

func TestRedisCache_Get(t *testing.T) {
	cache, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	testData := "test_value"

	// Set a test value (JSON encoded)
	jsonData, err := json.Marshal(testData)
	assert.NoError(t, err)
	err = mr.Set("test_key", string(jsonData))
	assert.NoError(t, err)

	// Test successful get
	var result string
	err = cache.Get(ctx, "test_key", &result)
	assert.NoError(t, err)
	assert.Equal(t, testData, result)

	// Test get non-existent key
	var notFound string
	err = cache.Get(ctx, "non_existent", &notFound)
	assert.Error(t, err)

	// Test get with invalid connection
	mr.Close() // Close Redis to simulate connection error
	err = cache.Get(ctx, "test_key", &result)
	assert.Error(t, err)
}

func TestRedisCache_Delete(t *testing.T) {
	cache, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	testData := "test_value"

	// Set a test value
	err := mr.Set("test_key", testData)
	assert.NoError(t, err)

	// Test successful delete
	err = cache.Delete(ctx, "test_key")
	assert.NoError(t, err)

	// Verify key was deleted
	exists := mr.Exists("test_key")
	assert.False(t, exists)
}

func TestRedisCache_Clear(t *testing.T) {
	cache, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Set multiple test values
	err := mr.Set("test_key1", "test_value1")
	assert.NoError(t, err)
	err = mr.Set("test_key2", "test_value2")
	assert.NoError(t, err)

	// Test successful clear
	err = cache.Clear(ctx)
	assert.NoError(t, err)

	// Verify all keys were deleted
	assert.Equal(t, 0, len(mr.Keys()))
} 