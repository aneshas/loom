package views

import (
	"github.com/a-h/templ"
	"github.com/aneshas/helloapp/web/views/layouts"
	"github.com/aneshas/loom"
	"github.com/labstack/echo/v4"
)

func Render(c echo.Context, component templ.Component, opts ...loom.RenderOption) error {
	options := loom.RenderOptions{
		Title: "Hello World",
	}

	for _, opt := range opts {
		options = opt(options)
	}

	if options.Layout == nil {
		options.Layout = layouts.App
	}

	return layouts.
		Root(options.Title, options.Layout(component)).
		Render(c.Request().Context(), c.Response().Writer)
}
