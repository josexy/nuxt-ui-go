package nui

// limitTextEdit limits inserted text while preserving the untouched suffix and
// counting UTF-16 code units, as HTML maxlength does. It runs after native input,
// leaving composition and selection handling with the editor.
func limitTextEdit(before, after string, limit int) string {
	units := func(s []rune) int {
		n := 0
		for _, r := range s {
			n++
			if r > 0xffff {
				n++
			}
		}
		return n
	}
	a, b := []rune(before), []rune(after)
	if limit <= 0 || units(b) <= limit {
		return after
	}
	start := 0
	for start < len(a) && start < len(b) && a[start] == b[start] {
		start++
	}
	endA, endB := len(a), len(b)
	for endA > start && endB > start && a[endA-1] == b[endB-1] {
		endA--
		endB--
	}
	remaining := max(0, limit-units(b[:start])-units(b[endB:]))
	end := start
	for end < endB {
		width := 1
		if b[end] > 0xffff {
			width = 2
		}
		if width > remaining {
			break
		}
		remaining -= width
		end++
	}
	return string(b[:end]) + string(b[endB:])
}
