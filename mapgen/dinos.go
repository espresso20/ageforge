package mapgen

// The dinosaur and asteroid rendering layer has been removed. The Primitive Age
// is kept grounded in human prehistory with clean prehistoric wilderness, foragers,
// flint knapping, and campfires.

// DinoHues is retained for compatibility and returns nil.
func DinoHues() []Hue {
	return nil
}

// paintDinos is a deprecated no-op.
func (f *frame) paintDinos() {}
