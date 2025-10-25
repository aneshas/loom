package loom

import (
	"net/url"

	"github.com/go-playground/form"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	*Deps

	FormDecoder *form.Decoder
	Validator   *validator.Validate
}

func (cont *Controller) BindForm(c echo.Context, dst any) (ViewModel, error) {
	req := c.Request()

	err := req.ParseForm()
	if err != nil {
		return ViewModel{}, err
	}

	err = cont.FormDecoder.Decode(dst, req.Form)
	if err != nil {
		return ViewModel{}, err
	}

	var validationErrors validator.ValidationErrors

	err = cont.Validator.Struct(dst)
	if err != nil {
		validationErrors = err.(validator.ValidationErrors)
	}

	model := cont.toViewModel(req.Form, validationErrors)

	return model, nil
}

func (cont *Controller) toViewModel(form url.Values, errors validator.ValidationErrors) ViewModel {
	m := ViewModel{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}

	for _, err := range errors {
		m.Errors[err.Field()] = err.Error()
	}

	for key, value := range form {
		m.Values[key] = value[0]
	}

	return m
}
