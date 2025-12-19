package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/response"
)

func errorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := utils.StatusMessage(code)

		if err != nil {
			if fiberErr, ok := err.(*fiber.Error); ok {
				code = fiberErr.Code
				if fiberErr.Message != "" {
					message = fiberErr.Message
				} else {
					message = utils.StatusMessage(code)
				}
			} else {
				message = err.Error()
			}
		}

		resp := response.Response{
			Code:    code,
			Message: message,
		}
		return c.Status(code).JSON(resp)
	}
}
