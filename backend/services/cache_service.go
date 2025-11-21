package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

// NewCacheService creates a new cache service with Redis
func NewCacheService() *CacheService {
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	cacheTTL, _ := strconv.Atoi(getEnv("CACHE_TTL", "86400")) // Default 24 hours

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       redisDB,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️  Redis connection failed: %v - Cache disabled", err)
		return &CacheService{
			client: nil,
			ctx:    ctx,
			ttl:    time.Duration(cacheTTL) * time.Second,
		}
	}

	log.Printf("✅ Redis connected successfully (TTL: %d seconds)", cacheTTL)

	return &CacheService{
		client: client,
		ctx:    ctx,
		ttl:    time.Duration(cacheTTL) * time.Second,
	}
}

// generateCacheKey generates a consistent cache key from input parameters
func (cs *CacheService) generateCacheKey(prefix string, params ...string) string {
	// Concatenate all params
	combined := prefix
	for _, param := range params {
		combined += ":" + param
	}

	// Create SHA256 hash for consistent short keys
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%s:%x", prefix, hash[:16]) // Use first 16 bytes of hash
}

// GetSEO retrieves cached SEO content
func (cs *CacheService) GetSEO(videoID string, clipStartTime, clipEndTime float64, clipTitle string) (*SEOContent, error) {
	if cs.client == nil {
		return nil, fmt.Errorf("cache disabled")
	}

	key := cs.generateCacheKey("seo",
		videoID,
		fmt.Sprintf("%.2f", clipStartTime),
		fmt.Sprintf("%.2f", clipEndTime),
		clipTitle,
	)

	val, err := cs.client.Get(cs.ctx, key).Result()
	if err == redis.Nil {
		log.Printf("🔍 [Cache MISS] SEO for %s (%.1f-%.1f)", videoID, clipStartTime, clipEndTime)
		return nil, fmt.Errorf("cache miss")
	} else if err != nil {
		log.Printf("❌ [Cache ERROR] Failed to get SEO: %v", err)
		return nil, err
	}

	var seo SEOContent
	if err := json.Unmarshal([]byte(val), &seo); err != nil {
		log.Printf("❌ [Cache ERROR] Failed to unmarshal SEO: %v", err)
		return nil, err
	}

	log.Printf("✅ [Cache HIT] SEO for %s (%.1f-%.1f)", videoID, clipStartTime, clipEndTime)
	return &seo, nil
}

// SetSEO stores SEO content in cache
func (cs *CacheService) SetSEO(videoID string, clipStartTime, clipEndTime float64, clipTitle string, seo *SEOContent) error {
	if cs.client == nil {
		return fmt.Errorf("cache disabled")
	}

	key := cs.generateCacheKey("seo",
		videoID,
		fmt.Sprintf("%.2f", clipStartTime),
		fmt.Sprintf("%.2f", clipEndTime),
		clipTitle,
	)

	data, err := json.Marshal(seo)
	if err != nil {
		log.Printf("❌ [Cache ERROR] Failed to marshal SEO: %v", err)
		return err
	}

	if err := cs.client.Set(cs.ctx, key, data, cs.ttl).Err(); err != nil {
		log.Printf("❌ [Cache ERROR] Failed to set SEO: %v", err)
		return err
	}

	log.Printf("💾 [Cache SET] SEO for %s (%.1f-%.1f) - TTL: %v", videoID, clipStartTime, clipEndTime, cs.ttl)
	return nil
}

// GetTranscript retrieves cached transcript
func (cs *CacheService) GetTranscript(videoID string) (string, error) {
	if cs.client == nil {
		return "", fmt.Errorf("cache disabled")
	}

	key := fmt.Sprintf("transcript:%s", videoID)

	val, err := cs.client.Get(cs.ctx, key).Result()
	if err == redis.Nil {
		log.Printf("🔍 [Cache MISS] Transcript for %s", videoID)
		return "", fmt.Errorf("cache miss")
	} else if err != nil {
		return "", err
	}

	log.Printf("✅ [Cache HIT] Transcript for %s", videoID)
	return val, nil
}

// SetTranscript stores transcript in cache
func (cs *CacheService) SetTranscript(videoID string, transcript string) error {
	if cs.client == nil {
		return fmt.Errorf("cache disabled")
	}

	key := fmt.Sprintf("transcript:%s", videoID)

	if err := cs.client.Set(cs.ctx, key, transcript, cs.ttl).Err(); err != nil {
		return err
	}

	log.Printf("💾 [Cache SET] Transcript for %s - TTL: %v", videoID, cs.ttl)
	return nil
}

// GetSuggestedClips retrieves cached suggested clips
func (cs *CacheService) GetSuggestedClips(videoID string) (string, error) {
	if cs.client == nil {
		return "", fmt.Errorf("cache disabled")
	}

	key := fmt.Sprintf("clips:%s", videoID)

	val, err := cs.client.Get(cs.ctx, key).Result()
	if err == redis.Nil {
		log.Printf("🔍 [Cache MISS] Suggested clips for %s", videoID)
		return "", fmt.Errorf("cache miss")
	} else if err != nil {
		return "", err
	}

	log.Printf("✅ [Cache HIT] Suggested clips for %s", videoID)
	return val, nil
}

// SetSuggestedClips stores suggested clips in cache
func (cs *CacheService) SetSuggestedClips(videoID string, clips string) error {
	if cs.client == nil {
		return fmt.Errorf("cache disabled")
	}

	key := fmt.Sprintf("clips:%s", videoID)

	if err := cs.client.Set(cs.ctx, key, clips, cs.ttl).Err(); err != nil {
		return err
	}

	log.Printf("💾 [Cache SET] Suggested clips for %s - TTL: %v", videoID, cs.ttl)
	return nil
}

// InvalidateVideo clears all cached data for a video
func (cs *CacheService) InvalidateVideo(videoID string) error {
	if cs.client == nil {
		return fmt.Errorf("cache disabled")
	}

	keys := []string{
		fmt.Sprintf("transcript:%s", videoID),
		fmt.Sprintf("clips:%s", videoID),
	}

	// Also invalidate SEO keys (we need to scan for them)
	pattern := fmt.Sprintf("seo:%s:*", videoID)
	iter := cs.client.Scan(cs.ctx, 0, pattern, 0).Iterator()
	for iter.Next(cs.ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := cs.client.Del(cs.ctx, keys...).Err(); err != nil {
			return err
		}
		log.Printf("🗑️  [Cache INVALIDATE] Cleared %d keys for video %s", len(keys), videoID)
	}

	return nil
}

// Close closes the Redis connection
func (cs *CacheService) Close() error {
	if cs.client != nil {
		return cs.client.Close()
	}
	return nil
}
