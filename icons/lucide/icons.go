// Package lucide provides the complete pinned Lucide SVG set through individually
// linkable, lazily parsed getters. Unused getters and resources can be discarded
// by the Go linker; importing the package does not parse the full collection.
// SVGs come from lucide-icons/lucide commit a630cec77eac6568b989e87e4947a8bfe45caacd.
// See LICENSE for Lucide and derived Feather notices.
package lucide

import (
	"sync"

	"github.com/egoist/mygo/ui"
)

type cached struct {
	once sync.Once
	svg  *ui.SVG
}

func (c *cached) load(source []byte) *ui.SVG {
	c.once.Do(func() { c.svg = ui.MustParseSVG(source) })
	return c.svg
}
