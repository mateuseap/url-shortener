package main

import (
	"crypto/sha256"
	"encoding/base64"
	"os"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

var (
	urlDB = make(map[string]string)
	mu    sync.Mutex
)

func helloWorld(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Hello World",
	})
}

func generateShortURL(longURL string) string {
	hash := sha256.Sum256([]byte(longURL))
	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}

func apiBaseURL(c fiber.Ctx) string {
	protocol := "http"
	if c.Protocol() == "https" {
		protocol = "https"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return protocol + "://" + c.Hostname() + ":" + port + "/"
}

func shortenURL(c fiber.Ctx) error {
	var req struct {
		URL string `json:"url"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	shortURL := generateShortURL(req.URL)

	mu.Lock()
	urlDB[shortURL] = req.URL
	mu.Unlock()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"short_url": apiBaseURL(c) + shortURL,
	})
}

func redirectURL(c fiber.Ctx) error {
	shortCode := c.Params("shortURL")

	mu.Lock()
	longURL, exists := urlDB[shortCode]
	mu.Unlock()

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Short URL not found",
		})
	}

	return c.Redirect().Status(fiber.StatusFound).To(longURL)
}

func main() {
	api := fiber.New()

	api.Use(logger.New(logger.Config{
		Format: "${time} [${ip}] ${status} - ${latency} ${method} ${path}\n",
	}))

	api.Get("/hello-world", helloWorld)
	api.Post("/shorten-url", shortenURL)
	api.Get("/:shortURL", redirectURL)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	api.Listen(":" + port)
}
