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
			"key":   "value",
			"hello": "world",
			"empty": "",
		},
	}
	k, v := options.getMetaData()
	assert.Len(t, k, 2)
	assert.Len(t, v, 2)
	assert.Contains(t, k, "key")
	assert.Contains(t, k, "hello")
	assert.Contains(t, v, "value")
	assert.Contains(t, v, "world")
}
