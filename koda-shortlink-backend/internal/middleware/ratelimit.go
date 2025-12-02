package middleware

import (
	"context"
	"fmt"
	"koda-shortlink-backend/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redisClient *redis.Client
	maxRequests int
	window      time.Duration
}

func NewRateLimiter(redisClient *redis.Client, maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		maxRequests: maxRequests,
		window:      window,
	}
}

func (r *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		key := fmt.Sprintf("ratelimit:%s", identifier)
		ctx := context.Background()

		count, err := r.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		if count >= r.maxRequests {
			response.Error(c, 429, "Rate limit exceeded. Please try again later.", nil)
			c.Abort()
			return
		}

		pipe := r.redisClient.Pipeline()
		pipe.Incr(ctx, key)
		if count == 0 {
			pipe.Expire(ctx, key, r.window)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		c.Next()
	}
}

func (r *RateLimiter) LimitByEndpoint(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		endpoint := c.FullPath()
		key := fmt.Sprintf("ratelimit:%s:%s", identifier, endpoint)
		ctx := context.Background()

		count, err := r.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		if count >= maxRequests {
			response.Error(c, 429, "Rate limit exceeded for this endpoint", nil)
			c.Abort()
			return
		}

		pipe := r.redisClient.Pipeline()
		pipe.Incr(ctx, key)
		if count == 0 {
			pipe.Expire(ctx, key, window)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		c.Next()
	}
}
