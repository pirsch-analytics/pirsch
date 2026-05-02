package channel

import (
	"fmt"
	"strings"
)

var searchChannel,
	socialChannel,
	shoppingChannel,
	videoChannel,
	aiChannel []string

func init() {
	for hostname, c := range channel {
		switch c {
		case "SOURCE_CATEGORY_SEARCH":
			searchChannel = append(searchChannel, strings.ToLower(hostname))
			break
		case "SOURCE_CATEGORY_SOCIAL":
			socialChannel = append(socialChannel, strings.ToLower(hostname))
			break
		case "SOURCE_CATEGORY_SHOPPING":
			shoppingChannel = append(shoppingChannel, strings.ToLower(hostname))
			break
		case "SOURCE_CATEGORY_VIDEO":
			videoChannel = append(videoChannel, strings.ToLower(hostname))
			break
		case "SOURCE_CATEGORY_AI":
			aiChannel = append(aiChannel, strings.ToLower(hostname))
			break
		default:
			panic(fmt.Sprintf("unknown channel type: %s", c))
		}
	}
}
