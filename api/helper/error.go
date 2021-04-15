package helper

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sahalazain/kafpro/api/presenter"
)

//Error set error
func Error(ctx *fiber.Ctx, err *fiber.Error, message string) error {
	hte := presenter.HTTPError{
		Code:    err.Code,
		Message: message,
	}

	ctx.Status(err.Code)
	return ctx.JSON(hte)
}
