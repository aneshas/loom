package web

import (
	"net/http"

	"github.com/aneshas/loom"
	"github.com/arl/statsviz"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// ConfigureServer is where we configure the echo server and global middleware
func ConfigureServer(g *loom.Loom) {
	g.E.HideBanner = true

	g.E.Use(
		middleware.CSRFWithConfig(middleware.CSRFConfig{
			TokenLookup: "form:_csrf",
		}),
		loom.CSRFMiddleware,
		loom.FlashMiddleware,
	)

	g.E.Static("/assets", "web/views/assets")

	g.E.File("/favicon.ico", "web/views/assets/favicon.ico")
	g.E.File("/robots.txt", "web/views/assets/robots.txt")

	mux := http.NewServeMux()

	statsviz.Register(mux)

	g.E.GET("/debug/statsviz/", echo.WrapHandler(mux))
	g.E.GET("/debug/statsviz/*", echo.WrapHandler(mux))
}
