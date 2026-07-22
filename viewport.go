package omnitui

// Viewport is the current terminal size in cells.
type Viewport struct {
	Width  int
	Height int
}

// UseViewport returns the terminal size for the current render.
func UseViewport(ctx Context) Viewport {
	instance := hookInstance(ctx, "UseViewport")
	return Viewport{Width: instance.app.width, Height: instance.app.height}
}
