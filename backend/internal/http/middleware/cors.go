package middleware

import (
	"strings"
	"time"

	"github.com/XATAB1CH/v2b/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS middleware using gin-contrib/cors (safe: does NOT read request body).
func CORS(cfg config.Config) gin.HandlerFunc {
	origins := normalizeOrigins(cfg.CORSAllowOrigins)

	c := cors.Config{
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:       12 * time.Hour,
	}

	// Special case: allow all
	if len(origins) == 1 && origins[0] == "*" {
		c.AllowAllOrigins = true
	} else {
		c.AllowOrigins = origins
	}

	return cors.New(c)
}

func normalizeOrigins(in []string) []string {
	if len(in) == 0 {
		return []string{"*"}
	}

	// if config already parsed list -> just trim
	out := make([]string, 0, len(in))
	for _, o := range in {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		out = append(out, o)
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}
