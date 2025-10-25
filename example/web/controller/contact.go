package controller

type Contact struct {
	Name  string  `form:"name" validate:"required"`
	Email string  `form:"email" validate:"required,email"`
	Phone *string `form:"phone"`
}
