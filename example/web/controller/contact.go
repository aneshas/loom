package controller

import (
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aneshas/helloapp/internal/db/model"
)

type Contact struct {
	ID        *int64    `json:"id" form:"id"`
	Email     string    `json:"email" form:"email" validate:"required"`
	Name      string    `json:"name" form:"name" validate:"required"`
	Phone     *string   `json:"phone" form:"phone"`
	CreatedAt time.Time `json:"created_at" form:"created_at"`
	UpdatedAt time.Time `json:"updated_at" form:"updated_at"`
}

// ContactToDB converts Contact to model.Contact
func (m *Contact) ToDB() *model.Contact {
	return &model.Contact{
		ID:        null.Int64FromPtr(m.ID),
		Email:     m.Email,
		Name:      m.Name,
		Phone:     null.StringFromPtr(m.Phone),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// ContactFromDB converts model.Contact to Contact
func ContactFromDB(m *model.Contact) *Contact {
	return &Contact{
		ID:        m.ID.Ptr(),
		Email:     m.Email,
		Name:      m.Name,
		Phone:     m.Phone.Ptr(),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
