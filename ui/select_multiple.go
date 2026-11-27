package nui

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
)

// SelectMultiple selects typed values from the state's submitted options. It
// shares Select's native virtual List path without converting business keys.
func SelectMultiple[T comparable](c *ui.Context, t Theme, selected *SelectionState[T], s *SelectState[T], o SelectOptions) ui.Element {
	if selected == nil || s == nil {
		panic("nui: SelectMultiple requires selection and state")
	}
	return selectControl(c, t, nil, selected, s, o)
}

func (s *SelectState[T]) updateSelectionLabel(selected *SelectionState[T]) {
	if s.labelSelection == selected && s.labelSelectionRevision == selected.Revision() && s.labelDataRevision == s.source.revision {
		return
	}
	labels := make([]string, 0, selected.Len())
	known := make(map[T]string, selected.Len())
	s.preferred = Optional[T]{}
	for _, key := range selected.values {
		label, found := s.labels[key]
		if at, exists := s.source.metadata.index[key]; exists {
			option := s.source.items[at]
			label = option.Label
			if !s.preferred.Set && !option.Disabled {
				s.preferred = Set(key)
			}
		} else if !found {
			label = fmt.Sprint(key)
		}
		known[key] = label
		if label != "" {
			labels = append(labels, label)
		}
	}
	s.labels, s.label = known, strings.Join(labels, ", ")
	s.labelSelection = selected
	s.labelSelectionRevision, s.labelDataRevision = selected.Revision(), s.source.revision
}
