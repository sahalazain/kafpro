package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/sahalazain/go-common/logger"
	"github.com/sahalazain/kafpro/api/helper"
	"github.com/sahalazain/kafpro/service"
)

func Send(app service.App) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		log := logger.GetLoggerContext(ctx, "handler", "Send")
		req := make(map[string]interface{})
		topic := c.Params("topic")
		key := c.Params("key")

		if err := c.BodyParser(&req); err != nil {
			log.WithError(err).Error("Error parsing request body")
			return helper.Error(c, fiber.ErrBadRequest, err.Error())
		}

		if err := app.SendMessage(ctx, topic, key, req); err != nil {
			log.WithError(err).Error("Error sending message")
			return helper.Error(c, fiber.ErrBadGateway, err.Error())
		}

		return c.SendString("OK")
	}
}

func Get(app service.App) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		log := logger.GetLoggerContext(ctx, "handler", "Get")
		topic := c.Params("topic")
		group := c.Params("group")
		max := c.Query("max")

		m := 1

		if max != "" {
			i, _ := strconv.Atoi(max)
			if i > 0 {
				m = i
			}
		}

		out, err := app.BulkRetrieve(ctx, topic, group, m)
		if err != nil {
			log.WithError(err).Error("Error retrieving message")
			return helper.Error(c, fiber.ErrBadGateway, err.Error())
		}

		return c.JSON(out)
	}
}
