package loom

import (
	"context"

	"github.com/labstack/echo/v4"
)

func CSRFMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.WithValue(c.Request().Context(), "csrf", c.Get("csrf"))
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
