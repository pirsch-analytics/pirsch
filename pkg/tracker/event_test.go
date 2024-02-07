package tracker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventOptions_validate(t *testing.T) {
	options := EventOptions{
		Name: " test",
	}
	options.validate()
	assert.Equal(t, "test", options.Name)
}

func TestEventOptions_getMetaData(t *testing.T) {
	options := EventOptions{
		Meta: map[string]string{
			"key":     "value",
			" hello ": " world ",
			"empty":   "",
			"":        "ignore",
		},
	}
	k, v := options.getMetaData([]string{"author", "key"}, []string{"John", "override"})
	assert.Len(t, k, 3)
	assert.Len(t, v, 3)
	assert.Contains(t, k, "author")
	assert.Contains(t, k, "key")
	assert.Contains(t, k, "hello")
	assert.Contains(t, v, "John")
	assert.Contains(t, v, "value")
	assert.Contains(t, v, "world")
}
