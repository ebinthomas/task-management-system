package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

type TestData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func setupTestRedis(t *testing.T) (*RedisCache, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}

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
	testData := TestData{Name: "test", Value: 123}

	// Test successful set
	err := cache.Set(ctx, "test_key", testData, 1*time.Minute)
	assert.NoError(t, err)
	assert.True(t, mr.Exists("test_key"))

	// Test set with nil value
	err = cache.Set(ctx, "nil_key", nil, 1*time.Minute)
	assert.Error(t, err)
}

func TestRedisCache_Get(t *testing.T) {
	cache, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 123}

	// Set test data
	err := cache.Set(ctx, "test_key", testData, 1*time.Minute)
	assert.NoError(t, err)

	// Test successful get
	var retrieved TestData
	err = cache.Get(ctx, "test_key", &retrieved)
	assert.NoError(t, err)
	assert.Equal(t, testData, retrieved)

	// Test get non-existent key
	var empty TestData
	err = cache.Get(ctx, "non_existent", &empty)
	assert.Error(t, err)

	// Test get with expired key
	mr.SetTTL("test_key", -1*time.Second)
	err = cache.Get(ctx, "test_key", &empty)
	assert.Error(t, err)
}

func TestRedisCache_Delete(t *testing.T) {
	cache, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 123}

	// Set test data
	err := cache.Set(ctx, "test_key", testData, 1*time.Minute)
	assert.NoError(t, err)

	// Test successful delete
	err = cache.Delete(ctx, "test_key")
	assert.NoError(t, err)
	assert.False(t, mr.Exists("test_key"))

	// Test delete non-existent key
	err = cache.Delete(ctx, "non_existent")
	assert.NoError(t, err) // Redis DEL returns success even if key doesn't exist
}

func TestRedisCache_Clear(t *testing.T) {
	cache, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 123}

	// Set multiple test keys
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("test_key_%d", i)
		err := cache.Set(ctx, key, testData, 1*time.Minute)
		assert.NoError(t, err)
	}

	// Test clear
	err := cache.Clear(ctx)
	assert.NoError(t, err)

	// Verify all keys are removed
	assert.Equal(t, 0, len(mr.Keys()))
} 