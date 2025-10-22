package controller

import (
	"github.com/aneshas/helloapp/web/views"
	"github.com/aneshas/helloapp/web/views/components"
	"github.com/aneshas/helloapp/web/views/contacts"
	"github.com/aneshas/loom"
	"github.com/labstack/echo/v4"
)

type ContactsController struct {
	loom.Controller
}

func (ctrl *ContactsController) New(c echo.Context) error {
	m := components.FormModel{
		Errors: map[string]string{
			"name": "This field is required",
		},
		Values: map[string]string{
			"email": "mail@example.com",
		},
	}

	return views.Render(c, contacts.Form(m))
}
