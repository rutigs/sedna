package redirect

import (
	fiber "github.com/gofiber/fiber/v2"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rutigs/sedna/pkg/redis"
)

// RedirectRoute - route for redirecting shortened URLs
func RedirectRoute(redisSvc *redis.RedisSvc, lruCache *lru.Cache) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		urlPath := c.Params("url")

		// Check the local cache
		res, exists := lruCache.Get(urlPath)
		if exists {
			return c.Redirect(res.(string))
		}

		// Check the remote cache
		redirectUrl, exists := redisSvc.Get(urlPath)
		if exists {
			_ = lruCache.Add(urlPath, redirectUrl)
			return c.Redirect(redirectUrl)
		}

		return c.Status(404).SendString("No redirect found!")
	}
}
