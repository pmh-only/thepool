package main

import (
	"crypto/rand"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func openWebserver() {
	port := getEnvDefault("WEBSERVER_PORT", "8080")
	sizeLimitMBRaw := getEnvDefault("WEBSERVER_SIZE_LIMIT_MB", "128")
	sizeLimitMB, err := strconv.ParseInt(sizeLimitMBRaw, 10, 64)

	if err != nil {
		log.Fatalf("WEBSERVER_SIZE_LIMIT_MB: %s is invaild\n", sizeLimitMBRaw)
	}

	app := fiber.New(fiber.Config{
		Prefork:   true,
		BodyLimit: int(sizeLimitMB) * 1024 * 1024,
	})

	app.Use(logger.New())

	app.Static("/", "./public")
	app.Static("/assets", "./node_modules")

	app.Get("/api/config", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"chunkSize": sizeLimitMB,
		})
	})

	app.Get("/api/collections/:id", func(c *fiber.Ctx) error {
		collectionId := c.Params("id")
		collection := getCollection(collectionId)

		if collection == nil {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "collection not found",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data":    collection,
		})
	})

	app.Post("/api/collections", func(c *fiber.Ctx) error {
		var payload Collection

		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "payload invaild",
			})
		}

		payload.CollectionId = rand.Text()[:10]

		createCollection(payload)
		return c.JSON(fiber.Map{
			"success": true,
			"id":      payload.CollectionId,
		})
	})

	app.Post("/api/chunks", func(c *fiber.Ctx) error {
		reader := c.Context().RequestBodyStream()
		chunkId := rand.Text()[:10]

		chunkSizeMB := int64(c.Context().Request.Header.ContentLength())/1024/1024 + 1

		if chunkSizeMB < 0 {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "content length not provided",
			})
		}

		err := uploadChunk(chunkId, chunkSizeMB, reader)
		if err != nil {
			deleteChunk([]string{chunkId})
		}

		createChunk(chunkId, chunkSizeMB)

		return c.JSON(fiber.Map{
			"success": true,
			"id":      chunkId,
		})
	})

	log.Fatal(app.Listen(":" + port))
}
