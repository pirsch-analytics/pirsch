package pirsch

import (
	"testing"
)

func TestRunAtMidnight(t *testing.T) {
	cancel := RunAtMidnight(func() {
		t.Fatal("Function must not be called")
	})
	cancel()
}

func TestNewTenantID(t *testing.T) {
	if NewTenantID(-1).Valid {
		t.Fatal("-1 must not be a valid tenant ID")
	}

	if NewTenantID(0).Valid {
		t.Fatal("0 must not be a valid tenant ID")
	}

	if !NewTenantID(42).Valid {
		t.Fatal("42 must be a valid tenant ID")
	}
}

func TestShortenString(t *testing.T) {
	out := shortenString("Hello World", 5)

	if out != "Hello" {
		t.Fatalf("String must have been shortened to 'Hello', but was: %v", out)
	}

	out = shortenString("Hello World", 50)

	if out != "Hello World" {
		t.Fatalf("String must not have been shortened, but was: %v", out)
	}
}

func TestContainsString(t *testing.T) {
	list := []string{"a", "b", "c", "d"}

	if containsString(list, "e") {
		t.Fatal("List must not contain string 'e'")
	}

	if !containsString(list, "c") {
		t.Fatal("List must contain string 'c'")
	}
}
