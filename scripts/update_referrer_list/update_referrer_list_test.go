package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeGroups(t *testing.T) {
	groups := make(map[string]string)
	addGroups(groups, list{
		"unknown": {
			"Tripadvisor": {
				Domains: []string{
					"tripadvisor.com",
					"tripadvisor.be",
					"www.tripadvisor.be",
				},
			},
		},
	})
	addGroups(groups, list{
		"search": {
			"Google": {
				Domains: []string{
					"Google.com",
				},
			},
		},
		"unknown": {
			"Tripadvisor": {
				Domains: []string{
					"tripadvisor.com",
					"tripadvisor.fr",
				},
			},
		},
	})
	assert.Len(t, groups, 4)
	assert.Equal(t, groups["tripadvisor.com"], "Tripadvisor")
	assert.Equal(t, groups["tripadvisor.fr"], "Tripadvisor")
	assert.Equal(t, groups["tripadvisor.be"], "Tripadvisor")
	assert.Equal(t, groups["google.com"], "Google")
}
