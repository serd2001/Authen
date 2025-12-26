package presenters

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	SUCCESS = 1
	FAIL    = 0
)

/*
Return success response
*/
func ResponseSuccess(data interface{}) fiber.Map {
	t := time.Now()
	return fiber.Map{
		"timestamp": t.Format("2006-01-02-15-04-05"),
		"status":    SUCCESS,
		"items":     data,
		"error":     nil,
	}
}
func ResponseSuccessListData(data interface{}, currentPage, currentPageTotalItem, totalPage int) fiber.Map {
	t := time.Now()
	return fiber.Map{
		"timestamp": t.Format("2006-01-02-15-04-05"),
		"status":    SUCCESS,
		"items": fiber.Map{
			"list_data": data,
			"pagination": fiber.Map{
				"current_page":            currentPage,
				"current_page_total_item": currentPageTotalItem,
				"total_page":              totalPage,
			},
		},
		"error": nil,
	}
}

// ResponseError returns error response
func ResponseError(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(fiber.Map{
		"timestamp": time.Now().Format("2006-01-02-15-04-05"),
		"status":    FAIL,
		"items":     nil,
		"error":     message,
	})
}
