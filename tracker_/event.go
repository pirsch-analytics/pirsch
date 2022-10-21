package tracker

// EventOptions are the metadata fields for an event.
type EventOptions struct {
	// Name is the name of the event (required).
	Name string

	// Duration is an optional duration that is used to calculate an average.
	Duration uint32

	// Meta are optional fields that can be used to break down events.
	Meta map[string]string
}

func (options *EventOptions) getMetaData() ([]string, []string) {
	keys, values := make([]string, 0, len(options.Meta)), make([]string, 0, len(options.Meta))

	for k, v := range options.Meta {
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
}
