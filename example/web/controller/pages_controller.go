package controller

import (
	"github.com/aneshas/helloapp/web/views"
	"github.com/aneshas/helloapp/web/views/pages"
	"github.com/aneshas/loom"
	"github.com/labstack/echo/v4"
)

type PagesController struct {
	loom.Controller
}

func (ctrl *PagesController) Home(c echo.Context) error {
	return views.Render(c, pages.Home())
}

func (ctrl *PagesController) NotFound(c echo.Context) error {
	return views.Render(c, pages.NotFound())
}
