package main

import "github.com/gofiber/fiber/v2"

func registerAppRoutes(webApp *fiber.App) {
	registerIndex(webApp)
}

func registerIndex(webApp *fiber.App) {
	webApp.Get("/", func(c *fiber.Ctx) error {
		return c.Render("home/index", nil)
	})
}
