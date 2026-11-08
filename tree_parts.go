package nui

import "github.com/egoist/mygo/ui"

func finishTreeItem(t Theme, slot Slot, e ui.Element, base PartStyle, tree, node Decoration) {
	base = mergeStyle(base, t.components[ComponentTree].UI[slot])
	base = mergeStyle(base, tree.UI[slot])
	base = mergeStyle(base, node.UI[slot])
	base.apply(e)
	if tree.Style != nil {
		tree.Style(slot, e)
	}
	if node.Style != nil {
		node.Style(slot, e)
	}
}
