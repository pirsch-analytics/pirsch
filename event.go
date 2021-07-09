package pirsch

// EventOptions are the options to save a new event.
// The name is required. All other fields are optional.
type EventOptions struct {
	Name     string
	Duration int
	Meta     map[string]string
}

func (options *EventOptions) getMetaData() ([]string, []string) {
	keys, values := make([]string, 0, len(options.Meta)), make([]string, 0, len(options.Meta))

	for k, v := range options.Meta {
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
}
