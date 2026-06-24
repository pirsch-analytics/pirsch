package screen

// Classes is a list of common screen size classifications.
var Classes = []Class{
	{5120, "UHD 5K"},
	{3840, "UHD 4K"},
	{2560, "WQHD"},
	{1920, "Full HD"},
	{1280, "HD"},
	{1024, "XL"},
	{800, "L"},
	{600, "M"},
	{415, "S"},
	{1, "XS"},
	{0, ""},
}

// Class is a screen size classification.
type Class struct {
	MinWidth uint16
	Class    string
}
