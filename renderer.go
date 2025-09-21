package loom

import "github.com/a-h/templ"

type RenderOptions struct {
	Title  string
	Layout func(component templ.Component) templ.Component
}

type RenderOption func(RenderOptions) RenderOptions

func WithTitle(title string) RenderOption {
	return func(options RenderOptions) RenderOptions {
		options.Title = title
		return options
	}
}

func WithLayout(layout func(_ templ.Component) templ.Component) RenderOption {
	return func(options RenderOptions) RenderOptions {
		options.Layout = layout
		return options
	}
}
