package web

import (
	"github.com/aneshas/loom"
)

// ConfigureRoutes is where we configure the routes and handler middleware
func ConfigureRoutes(l *loom.Loom) {
	l.GET("/", "pages.home")
	l.GET("/contacts/new", "contacts.new")
	l.GET("*", "pages.not_found")
}
