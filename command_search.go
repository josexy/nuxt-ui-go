package nui

import "strings"

type commandFields uint8

const commandDefaultFields commandFields = 15

func commandFieldMask(fields []string) commandFields {
	if len(fields) == 0 {
		return commandDefaultFields
	}
	var mask commandFields
	for _, field := range fields {
		switch field {
		case "label":
			mask |= 1
		case "description":
			mask |= 2
		case "suffix":
			mask |= 4
		case "keywords":
			mask |= 8
		default:
			panic("nui: unsupported CommandPalette filter field")
		}
	}
	return mask
}

func matchCommand(it indexedCommand, query []rune, scratch [4][]int, fields commandFields, ignore bool) ([4][]int, int, bool) {
	var marks [4][]int
	if ignore || len(query) == 0 {
		return marks, 0, true
	}
	texts := [4][]rune{it.label, it.description, it.suffix, it.keywords}
	penalty := [4]int{0, 10, 5, 20}
	best, found := 0, false
	for i, text := range texts {
		if fields&(1<<i) == 0 {
			continue
		}
		positions, score, ok := fuzzy(text, query, scratch[i])
		if !ok {
			continue
		}
		marks[i] = positions
		score -= penalty[i]
		if !found || score > best {
			best, found = score, true
		}
	}
	return marks, best, found
}

// Group ranked results contiguously, preserving the first group's appearance
// unless the caller asks for the index's original group order.
func groupCommands(matches []CommandMatch, index *CommandIndex, parent string, preserve bool) []CommandMatch {
	groups := map[string][]CommandMatch{}
	order := []string{}
	for _, match := range matches {
		id := match.Action.Group
		if _, ok := groups[id]; !ok {
			order = append(order, id)
		}
		groups[id] = append(groups[id], match)
	}
	if preserve && index != nil {
		original := []string{}
		seen := map[string]bool{}
		for _, item := range index.levels[parent] {
			id := item.action.Group
			if !seen[id] && len(groups[id]) > 0 {
				original = append(original, id)
				seen[id] = true
			}
		}
		for _, id := range order {
			if !seen[id] {
				original = append(original, id)
			}
		}
		order = original
	}
	out := make([]CommandMatch, 0, len(matches))
	for _, id := range order {
		out = append(out, groups[id]...)
	}
	return out
}

func commandLabel(a CommandAction) string {
	parts := []string{}
	for _, value := range []string{a.Prefix, a.Label, a.Suffix} {
		if value != "" {
			parts = append(parts, value)
		}
	}
	return strings.Join(parts, " ")
}
