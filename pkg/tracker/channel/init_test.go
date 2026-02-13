package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	assert.NotZero(t, len(searchChannel))
	assert.NotZero(t, len(socialChannel))
	assert.NotZero(t, len(shoppingChannel))
	assert.NotZero(t, len(videoChannel))
	assert.NotZero(t, len(aiChannel))
}
