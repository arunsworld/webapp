package main

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func setupStatic(webApp *fiber.App, staticContent fs.FS) {
	webApp.Get("/static/:resource", func(c *fiber.Ctx) error {
		resourceName := c.Params("resource")

		f, err := staticContent.Open(resourceName)
		if err != nil {
			log.Warn().Err(err).Str("resource", resourceName).Msg("setupStatic: error opening resource")
			return c.Status(http.StatusNotFound).SendString("resource not found")
		}
		defer f.Close()

		switch {
		case strings.HasSuffix(resourceName, "css"):
			c.Set(fiber.HeaderContentType, "text/css")
		case strings.HasSuffix(resourceName, "js"):
			c.Set(fiber.HeaderContentType, fiber.MIMETextJavaScript)
		case strings.HasSuffix(resourceName, "jpeg"):
			c.Set(fiber.HeaderContentType, "image/jpeg")
		case strings.HasSuffix(resourceName, "png"):
			c.Set(fiber.HeaderContentType, "image/png")
		}

		return c.Status(http.StatusOK).SendStream(f)
	})
}
