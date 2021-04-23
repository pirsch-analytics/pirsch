package hit

type screenClass struct {
	minWidth int
	class    string
}

// ScreenClasses is a list of typical screen sizes used to group resolutions.
// Everything below is considered "XS" (tiny).
var ScreenClasses = []screenClass{
	{1440, "XXL"},
	{1024, "XL"},
	{800, "L"},
	{600, "M"},
	{415, "S"},
}

// GetScreenClass returns the screen class for given width in pixels.
func GetScreenClass(width int) string {
	if width <= 0 {
		return ""
	}

	for _, class := range ScreenClasses {
		if width >= class.minWidth {
			return class.class
		}
	}

	return "XS"
}
