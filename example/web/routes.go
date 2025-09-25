package web

import (
	"github.com/aneshas/loom"
)

// ConfigureRoutes is where we configure the routes and handler middleware
func ConfigureRoutes(l *loom.Loom) {
	// We could configure additional deps and or middleware per route
	// here by constructing them from g.Deps

	l.GET("/", "pages.home") // TODO better errors when not found (also render error page)
	l.GET("*", "pages.not_found")
}
