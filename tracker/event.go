package tracker

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

func (options *EventOptions) getMetaData() ([]string, []string) {
	keys, values := make([]string, 0, len(options.Meta)), make([]string, 0, len(options.Meta))

	for k, v := range options.Meta {
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
}
