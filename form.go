package loom

import (
	"context"

	"github.com/labstack/echo/v4"
)

func WithCSRFProtection(ctx context.Context, c echo.Context) context.Context {
	return context.WithValue(ctx, "csrf", c.Get("csrf"))
}
