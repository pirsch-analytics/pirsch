package pirsch

import "testing"

func TestParse(t *testing.T) {
	input := []string{
		// empty
		"",
		"  ",
		"'  '",
		` "   "`,

		// clean and simple
		"(system)",
		"version",
		"(system) version",

		// whitespace
		"   (system)   ",
		"   version    ",
		"   (   system   )   version   ",
		"   (  ;  system    ;  )   version   ",

		// multiple system entries and versions
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
	}
	expected := [][][]string{
		{{}, {}},
		{{}, {}},
		{{}, {}},
		{{}, {}},
		{{"system"}, {}},
		{{}, {"version"}},
		{{"system"}, {"version"}},
		{{"system"}, {}},
		{{}, {"version"}},
		{{"system"}, {"version"}},
		{{"system"}, {"version"}},
		{{"Macintosh", "Intel Mac OS X 10_10_5"}, {"Mozilla/5.0", "AppleWebKit/537.36", "(KHTML,", "like", "Gecko)", "Chrome/63.0.3239.132", "Safari/537.36"}},
	}

	for i, in := range input {
		system, versions := parseUserAgent(in)

		if !testStringSlicesEqual(system, expected[i][0]) || !testStringSlicesEqual(versions, expected[i][1]) {
			t.Fatalf("%v, expected: %v %v, was: %v %v", in, expected[i][0], expected[i][1], system, versions)
		}
	}
}

func testStringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
