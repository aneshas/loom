package controller

import (
	"database/sql"
	"net/http"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aneshas/helloapp/internal/db/model"
	"github.com/aneshas/helloapp/web/views"
	"github.com/aneshas/helloapp/web/views/contacts"
	"github.com/aneshas/loom"
	"github.com/labstack/echo/v4"
)

type ContactsController struct {
	loom.Controller

	db *sql.DB
}

func (ctrl *ContactsController) Init() error {
	ctrl.db = loom.MustGet[*sql.DB](ctrl.Deps)

	return nil
}

func (ctrl *ContactsController) New(c echo.Context) error {
	var m loom.ViewModel

	return views.Render(c, contacts.Form(m))
}

func (ctrl *ContactsController) Post(c echo.Context) error {
	var contact Contact

	m, err := ctrl.BindForm(c, &contact)
	if err != nil {
		return err
	}

	if m.HasErrors() {
		loom.FlashErrorNow(c, "Please fix form errors and re-submit.")
		return views.Render(c, contacts.Form(m))
	}

	dbContact := model.Contact{
		Name:  contact.Name,
		Email: contact.Email,
		Phone: null.StringFromPtr(contact.Phone),
	}

	err = dbContact.Insert(c.Request().Context(), ctrl.db, boil.Infer())
	if err != nil {
		loom.FlashErrorNow(c, err.Error()) // generic message but log the error for debugging
		return views.Render(c, contacts.Form(m))
	}

	loom.FlashSuccess(c, "Contact saved successfully!")

	return c.Redirect(http.StatusFound, "/contacts")
}
