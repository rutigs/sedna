package shortener

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	lru "github.com/hashicorp/golang-lru"
	"github.com/tidwall/gjson"

	"github.com/rutigs/sedna/pkg/redis"
)

// ShortenerRoute - route for creating shortened URLs
func ShortenerRoute(redisSvc *redis.RedisSvc, lruCache *lru.Cache) func(c *fiber.Ctx) error {
	// Required to seed a random starting string
	rand.Seed(time.Now().UnixNano())

	return func(c *fiber.Ctx) error {
		rawURL := gjson.GetBytes(c.Body(), "url").String()
		if !gjson.GetBytes(c.Body(), "url").Exists() {
			return c.Status(400).JSON(&fiber.Map{
				"error": "missing url in request body",
			})
		}

		url, err := url.Parse(rawURL)
		if err != nil {
			log.Println(err)
			return c.Status(400).JSON(&fiber.Map{
				"error": err,
			})
		}

		if err := validateUrl(url); err != nil {
			log.Println(err)
			return c.Status(400).JSON(&fiber.Map{
				"error": err,
			})
		}

		shortened, exists := redisSvc.Get(rawURL)
		if exists {
			_ = lruCache.Add(shortened, rawURL)
			return c.Status(200).JSON(&fiber.Map{
				rawURL: "/" + shortened,
			})
		}

		// Generate a unique path for this url
		tinyUrlPath := generateTinyUrlPath()
		for lruCache.Contains(tinyUrlPath) {
			tinyUrlPath = generateTinyUrlPath()
		}

		// Add both directions as key/vals
		// So if a new instance is brought up, it can refresh its in-mem cache
		// with existing redirects as well as detects duplicates in this method
		if urlToPath := redisSvc.Set(rawURL, tinyUrlPath); !urlToPath {
			return c.Status(500).JSON(&fiber.Map{
				"error": err,
			})
		}
		if pathToUrl := redisSvc.Set(tinyUrlPath, rawURL); !pathToUrl {
			return c.Status(500).JSON(&fiber.Map{
				"error": err,
			})
		}

		// Now add to the local in-mem cache for redirects
		_ = lruCache.Add(tinyUrlPath, rawURL)

		return c.Status(201).JSON(&fiber.Map{
			rawURL: "/" + tinyUrlPath,
		})
	}
}

// validateUrl - check if requested url is valid to be redirected
// Scheme must be http or https for this application spec
// Host must not be blank to be a valid url
func validateUrl(u *url.URL) error {
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "" {
		return fmt.Errorf("Invalid URL scheme for %s", u.String())
	}

	if u.Host == "" {
		return fmt.Errorf("Invalid URL host for %s", u.String())
	}

	return nil
}

var validCharacters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// generateTinyUrlPath - generates a random string of 7 characters
// 62 p 7 is +2.4T options
// incrementing an alphanumeric number would be more efficient but is more time consuming to develop
// however this alternate approach would make urls predictable which may or may not be a bad thing
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go/22892986#22892986
func generateTinyUrlPath() string {
	b := make([]rune, 7)
	for i := range b {
		b[i] = validCharacters[rand.Intn(62)]
	}
	return string(b)
}
