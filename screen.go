package pirsch

type screenClass struct {
	minWidth int
	class    string
}

// ScreenClasses is a list of typical screen sizes used to group resolutions.
// Everything below is considered "tiny".
var ScreenClasses = []screenClass{
	{1440, "Extra Extra Large"},
	{1024, "Extra Large"},
	{800, "Large"},
	{600, "Medium"},
	{415, "Small"},
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

	return "Tiny"
}
