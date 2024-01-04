package tracker

import "strings"

// EventOptions are the options to save a new event.
// The name is required. All other fields are optional.
type EventOptions struct {
	// Name is the name of the event (required).
	Name string

	// Duration is an optional duration that is used to calculate an average time on the dashboard.
	Duration uint32

	// Meta are optional fields used to break down the events that were send for a name.
	Meta map[string]string
}

func (options *EventOptions) validate() {
	options.Name = strings.TrimSpace(options.Name)
}

func (options *EventOptions) getMetaData(tagKeys, tagValues []string) ([]string, []string) {
	meta := make(map[string]string)

	for i := 0; i < len(tagKeys); i++ {
		meta[strings.TrimSpace(tagKeys[i])] = strings.TrimSpace(tagValues[i])
	}

	for k, v := range options.Meta {
		meta[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}

	keys, values := make([]string, 0, len(meta)), make([]string, 0, len(meta))

	for k, v := range meta {
		if k != "" && v != "" {
			keys = append(keys, k)
			values = append(values, v)
		}
	}

	return keys, values
}
