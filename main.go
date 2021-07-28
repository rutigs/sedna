package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/rutigs/sedna/pkg/redirect"
	"github.com/rutigs/sedna/pkg/redis"
	"github.com/rutigs/sedna/pkg/shortener"

	fiber "github.com/gofiber/fiber/v2"
	lru "github.com/hashicorp/golang-lru"
)

var (
	// Address of the backing redis store
	redisAddr string

	// Default LRU cache size for shortened URLs
	cacheSize int
)

func init() {
	// Override this when running locally
	flag.StringVar(&redisAddr, "redis", "redis:6379", "address of backing redis instance")
	flag.IntVar(&cacheSize, "cache-size", 8096, "LRU cache size for shortened URLs")

	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	// Create the backing redis store for tracking which URLs we have shortened -> shortened paths
	// LRU in memory cache for handling redirects quickly
	redisSvc := redis.NewRedisSvc(redisAddr)
	lruCache, err := lru.New(cacheSize)
	if err != nil {
		log.Println("Unable to create shortened url cache")
		os.Exit(1)
	}

	app := fiber.New()

	// Create api and v1 groups
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Register routes for shortening and redirecting
	v1.Post("/shorten", shortener.ShortenerRoute(redisSvc, lruCache))
	app.Get("/:url", redirect.RedirectRoute(redisSvc, lruCache))

	// Setup the channel to handle graceful shutdown signal
	shutdownSignalCh := make(chan os.Signal, 1)
	signal.Notify(shutdownSignalCh, os.Interrupt)

	go func() {
		_ = <-shutdownSignalCh
		log.Println("Gracefully shutting down the service...")
		_ = app.Shutdown()
	}()

	if err := app.Listen(":3000"); err != nil {
		log.Println(err)
	}
}
