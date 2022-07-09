package screen

type class struct {
	minWidth uint16
	class    string
}

// Classes is a list of typical screen sizes used to group resolutions.
// Everything below is considered "XS" (tiny).
var Classes = []class{
	{5120, "UHD 5K"},
	{3840, "UHD 4K"},
	{2560, "WQHD"},
	{1920, "Full HD"},
	{1280, "HD"},
	{1024, "XL"},
	{800, "L"},
	{600, "M"},
	{415, "S"},
}

// GetClass returns the screen class for given width in pixels.
func GetClass(width uint16) string {
	if width <= 0 {
		return ""
	}

	for _, class := range Classes {
		if width >= class.minWidth {
			return class.class
		}
	}

	return "XS"
}
