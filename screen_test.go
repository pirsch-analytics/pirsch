package pirsch

import "testing"

func TestGetScreenClass(t *testing.T) {
	if out := GetScreenClass(0); out != "" {
		t.Fatalf("No screen class must have been returned for 0 width, but was: %v", out)
	}

	if out := GetScreenClass(42); out != "XS" {
		t.Fatalf("Tiny screen class must have been returned, but was: %v", out)
	}

	if out := GetScreenClass(1024); out != "XL" {
		t.Fatalf("Large screen class must have been returned, but was: %v", out)
	}

	if out := GetScreenClass(1025); out != "XL" {
		t.Fatalf("Large screen class must have been returned, but was: %v", out)
	}

	if out := GetScreenClass(1919); out != "XXL" {
		t.Fatalf("Large screen class must have been returned, but was: %v", out)
	}
}
