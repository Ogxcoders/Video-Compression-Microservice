package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" {
			c.Next()
			return
		}

		providedKey := c.GetHeader("X-API-Key")
		if providedKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key is required",
			})
			c.Abort()
			return
		}

		if providedKey != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func DomainWhitelist(allowedDomains []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(allowedDomains) == 0 {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = c.GetHeader("Referer")
		}

		if origin == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Origin or Referer header is required",
			})
			c.Abort()
			return
		}

		allowed := false
		for _, domain := range allowedDomains {
			if strings.HasPrefix(origin, domain) {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Domain not allowed",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func CORS(allowedDomains []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin != "" && len(allowedDomains) > 0 {
			for _, domain := range allowedDomains {
				if strings.HasPrefix(origin, domain) {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, Authorization")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
