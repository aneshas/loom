package controller

import "github.com/aneshas/loom"

func Register(l *loom.Loom) {
	loom.Register[*PagesController](l)
}
