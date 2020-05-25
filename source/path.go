package source

import (
	"reflect"

	"github.com/derry6/vade-go/source/parser"
)

type configValue struct {
	store *pathStore
	value interface{}
}

type pathStore struct {
	path   string
	pri    int
	parser parser.Parser
	values map[string]interface{}
}

type pathHighToLow []*pathStore

func (n pathHighToLow) Len() int           { return len(n) }
func (n pathHighToLow) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n pathHighToLow) Less(i, j int) bool { return n[i].pri > n[j].pri }

func (n *pathStore) Update(values map[string]interface{}, withDeleted bool) (events []*Event) {
	if len(values) == 0 {
		for key, value := range n.values {
			if withDeleted {
				ev := NewEvent(Deleted, key)
				ev.ValueFrom = value
				events = append(events, ev)
			} else {
				values[key] = value
			}
		}
		return events
	}
	if len(n.values) == 0 {
		for key, value := range values {
			ev := NewEvent(Created, key)
			ev.ValueTo = value
			events = append(events, ev)
		}
		return events
	}
	for key, valueFrom := range n.values {
		if valueTo, ok := values[key]; ok {
			// updated
			if !reflect.DeepEqual(valueFrom, valueTo) {
				ev := NewEvent(Updated, key)
				ev.ValueFrom = valueFrom
				ev.ValueTo = valueTo
				events = append(events, ev)
			}
		} else {
			if withDeleted {
				ev := NewEvent(Deleted, key)
				ev.ValueFrom = valueFrom
				ev.ValueTo = valueTo
				events = append(events, ev)
			} else {
				values[key] = valueFrom
			}
		}
	}
	for key, valueTo := range values {
		if _, ok := n.values[key]; !ok {
			ev := NewEvent(Created, key)
			ev.ValueTo = valueTo
			events = append(events, ev)
		}
	}
	return events
}
