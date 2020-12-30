package pirsch

import "testing"

func TestGetScreenClass(t *testing.T) {
	if out := GetScreenClass(0); out != "" {
		t.Fatalf("No screen class must have been returned for 0 width, but was: %v", out)
	}

	if out := GetScreenClass(1024); out != "Extra Large" {
		t.Fatalf("Large screen class must have been returned, but was: %v", out)
	}

	if out := GetScreenClass(1025); out != "Extra Large" {
		t.Fatalf("Large screen class must have been returned, but was: %v", out)
	}

	if out := GetScreenClass(1919); out != "Extra Extra Large" {
		t.Fatalf("Large screen class must have been returned, but was: %v", out)
	}
}
