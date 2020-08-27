package pirsch

import "testing"

func TestFilter_Days(t *testing.T) {
	filter := NewFilter(NullTenant)

	// the default filter covers the past week NOT including today
	if days := filter.Days(); days != 6 {
		t.Fatalf("Filter must cover 6 days, but was: %v", days)
	}

	filter.From = pastDay(20)
	filter.To = today()
	filter.validate()

	if days := filter.Days(); days != 20 {
		t.Fatalf("Filter must cover 20 days, but was: %v", days)
	}
}

func TestFilter_Validate(t *testing.T) {
	filter := NewFilter(NullTenant)
	filter.validate()

	if filter == nil || !filter.From.Equal(pastDay(6)) || !filter.To.Equal(pastDay(0)) {
		t.Fatalf("Filter not as expected: %v", filter)
	}

	filter = &Filter{From: pastDay(2), To: pastDay(5)}
	filter.validate()

	if !filter.From.Equal(pastDay(5)) || !filter.To.Equal(pastDay(2)) {
		t.Fatalf("Filter not as expected: %v", filter)
	}
}
