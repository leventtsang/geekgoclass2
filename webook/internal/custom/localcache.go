package custom

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

// CodeCache 接口定义
type CodeCache interface {
	GetCode(key string) (string, error)
	SetCode(key string, value string, expiration time.Duration) error
}

// CodeRedisCache 原有的基于Redis的接口实现（代码省略）
type CodeRedisCache struct {
	// Redis 客户端等
}

func (c *CodeRedisCache) GetCode(key string) (string, error) {
	// 实现代码
	return "", nil
}

func (c *CodeRedisCache) SetCode(key string, value string, expiration time.Duration) error {
	// 实现代码
	return nil
}

// CodeLocalCache 基于本地的接口实现
type CodeLocalCache struct {
	Mu     sync.RWMutex
	Cache  map[string]string // 注意这里改为了大写开头
	Expiry map[string]time.Time
}

func NewCodeLocalCache() *CodeLocalCache {
	return &CodeLocalCache{
		Cache:  make(map[string]string),    // 改为大写
		Expiry: make(map[string]time.Time), // 改为大写
	}
}

func (c *CodeLocalCache) GetCode(key string) (string, error) {
	c.Mu.RLock()         // 改为大写
	defer c.Mu.RUnlock() // 改为大写

	if value, found := c.Cache[key]; found { // 改为大写
		if expireTime, exists := c.Expiry[key]; exists && time.Now().Before(expireTime) { // 改为大写
			return value, nil
		}
		// 清理过期的缓存
		delete(c.Cache, key)  // 改为大写
		delete(c.Expiry, key) // 改为大写
	}
	return "", nil
}

func (c *CodeLocalCache) SetCode(key string, value string, expiration time.Duration) error {
	c.Mu.Lock()         // 改为大写
	defer c.Mu.Unlock() // 改为大写

	c.Cache[key] = value // 改为大写
	if expiration > 0 {
		c.Expiry[key] = time.Now().Add(expiration) // 改为大写
	}
	return nil
}

// NewLocalRateLimitMiddleware 代码
func NewLocalRateLimitMiddleware(codeCache CodeCache, limit int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() // 使用客户端IP作为键

		if cnt, _ := codeCache.GetCode(key); cnt != "" {
			// 实现你的速率限制逻辑，如果超过限制则返回错误
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "Too Many Requests",
			})
			return
		}

		// 实现将请求计数存入缓存的逻辑
		codeCache.SetCode(key, "1", duration)

		c.Next()
	}
}
