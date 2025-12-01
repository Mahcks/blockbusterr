package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// RequestLogger logs incoming requests with timing information
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Continue to next handler
		err := c.Next()

		// Log the request
		duration := time.Since(start)
		log.Infof("[%s] %s - %d (%s)",
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			duration,
		)

		return err
	}
}

// SecureHeaders adds security headers to responses
func SecureHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Add security headers
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		return c.Next()
	}
}

// ValidateContentType validates request content type for POST/PUT/PATCH
func ValidateContentType() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only check for methods that typically have a body
		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
			contentType := c.Get("Content-Type")

			// Allow form data and JSON
			if contentType != "application/x-www-form-urlencoded" &&
				contentType != "application/json" &&
				contentType != "multipart/form-data" {
				return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
					"error": "Unsupported content type. Use application/x-www-form-urlencoded or application/json",
				})
			}
		}

		return c.Next()
	}
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BanDuration       time.Duration
}

// SimpleRateLimit implements a basic rate limiter using in-memory storage
func SimpleRateLimit(config RateLimitConfig) fiber.Handler {
	type visitor struct {
		count       int
		lastSeen    time.Time
		banned      bool
		bannedUntil time.Time
	}

	visitors := make(map[string]*visitor)

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()

		v, exists := visitors[ip]
		if !exists {
			v = &visitor{count: 0, lastSeen: now}
			visitors[ip] = v
		}

		// Check if banned
		if v.banned && now.Before(v.bannedUntil) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests. Please try again later.",
			})
		} else if v.banned && now.After(v.bannedUntil) {
			// Unban
			v.banned = false
			v.count = 0
		}

		// Reset count if more than 1 minute has passed
		if now.Sub(v.lastSeen) > time.Minute {
			v.count = 0
		}

		v.count++
		v.lastSeen = now

		// Check if over limit
		if v.count > config.RequestsPerMinute {
			v.banned = true
			v.bannedUntil = now.Add(config.BanDuration)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded. Please try again later.",
			})
		}

		return c.Next()
	}
}
