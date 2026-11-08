package showcase

//go:generate go run ../cmd/genshowcase -root ../..

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type enumValue struct {
	Name  string
	Value reflect.Value
}

// ValueSchema describes a public Go value without exposing private widget state.
type ValueSchema struct {
	Type      string       `json:"type"`
	Kind      string       `json:"kind"`
	Item      string       `json:"item,omitempty"`
	Fields    []ValueField `json:"fields,omitempty"`
	Choices   []string     `json:"choices,omitempty"`
	Length    int          `json:"length,omitempty"`
	Sample    any          `json:"sample"`
	Returns   []string     `json:"returns,omitempty"`
	Arguments []string     `json:"arguments,omitempty"`
}

type ValueField struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Description Text            `json:"description"`
	Control     string          `json:"control,omitempty"`
	Choices     []string        `json:"choices,omitempty"`
	Hint        *Text           `json:"hint,omitempty"`
	When        *FieldCondition `json:"when,omitempty"`
	Slots       []string        `json:"slots,omitempty"`
}

type FieldCondition struct {
	Field string `json:"field"`
	Value string `json:"value"`
	Label Text   `json:"label"`
}

type Binding struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	Fields     []ValueField `json:"fields,omitempty"`
	Defaults   any          `json:"defaults"`
	typ        reflect.Type
	start, end int
	literal    ast.Expr
}

func goType(t reflect.Type) string {
	if t.PkgPath() == "github.com/josexy/nuxt-ui-go" {
		return "nui." + strings.NewReplacer("github.com/josexy/nuxt-ui-go.", "nui.", "github.com/egoist/mygo/ui.", "ui.").Replace(t.Name())
	}
	if t.Name() != "" {
		return t.String()
	}
	switch t.Kind() {
	case reflect.Pointer:
		return "*" + goType(t.Elem())
	case reflect.Slice:
		return "[]" + goType(t.Elem())
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), goType(t.Elem()))
	case reflect.Map:
		return "map[" + goType(t.Key()) + "]" + goType(t.Elem())
	case reflect.Func:
		var in, out []string
		for i := 0; i < t.NumIn(); i++ {
			in = append(in, goType(t.In(i)))
		}
		for i := 0; i < t.NumOut(); i++ {
			out = append(out, goType(t.Out(i)))
		}
		result := "func(" + strings.Join(in, ", ") + ")"
		if len(out) == 1 {
			result += " " + out[0]
		} else if len(out) > 1 {
			result += " (" + strings.Join(out, ", ") + ")"
		}
		return result
	}
	return t.String()
}

func optional(t reflect.Type) bool {
	return t.Kind() == reflect.Struct && strings.HasPrefix(t.Name(), "Optional[") && t.PkgPath() == "github.com/josexy/nuxt-ui-go"
}

func publicFields(t reflect.Type) []reflect.StructField {
	var result []reflect.StructField
	if t.Kind() != reflect.Struct {
		return result
	}
	for _, field := range reflect.VisibleFields(t) {
		if field.IsExported() && !field.Anonymous {
			result = append(result, field)
		}
	}
	return result
}

func fieldSchemas(t reflect.Type) []ValueField {
	notes := APINotes()
	var fields []ValueField
	for _, field := range publicFields(t) {
		choices := enumNames(field.Type, t.Name(), field.Name)
		if field.Name == "FilterFields" {
			choices = []string{"label", "description", "value"}
			if t.Name() == "SelectMenuOptions" || t.Name() == "ListboxOptions" {
				choices = append(choices, "id", "group")
			} else if t.Name() == "MenuOptions" || t.Name() == "MenuContentOptions" {
				choices = []string{"label", "description", "id", "shortcut"}
			} else if t.Name() == "CommandPaletteOptions" {
				choices = []string{"label", "description", "suffix", "keywords"}
			}
		}
		if field.Type.Kind() == reflect.Func {
			choices = fieldFunctionModes(field.Type, field.Name)
		}
		description := notes[field.Name]
		owner, _, _ := strings.Cut(t.Name(), "[")
		if specific, ok := notes[owner+"."+field.Name]; ok {
			description = specific
		}
		editor := ValueField{Name: field.Name, Type: goType(field.Type), Description: description, Choices: choices}
		if t == reflect.TypeFor[nui.CalendarOptions]() {
			switch field.Name {
			case "UI", "Style":
				editor.Slots = []string{"root", "header", "heading", "body", "grid", "gridRow", "weekDay", "weekNumber", "item"}
			case "MaximumDays", "FixedDate", "AllowNonContiguousRanges", "IsDateHighlightable":
				editor.When = &FieldCondition{Field: "Range", Value: "true", Label: Text{ZH: "仅范围选择", EN: "Range selection only"}}
			case "ViewControlOptions":
				editor.Hint = &Text{ZH: "仅 ViewControl 启用时显示", EN: "Shown while ViewControl is enabled"}
			case "YearsPerPage", "IsYearDisabled", "IsYearUnavailable":
				editor.Hint = &Text{ZH: "用于年导航视图", EN: "Used by the year navigation view"}
			case "IsMonthDisabled", "IsMonthUnavailable":
				editor.Hint = &Text{ZH: "用于月导航视图", EN: "Used by the month navigation view"}
			}
		}
		if t == reflect.TypeFor[nui.TextareaOptions]() {
			switch field.Name {
			case "Password":
				editor.Hint = &Text{ZH: "多行输入不适用", EN: "Not used by multiline input"}
			case "OnSubmit":
				editor.Hint = &Text{ZH: "Enter 用于换行", EN: "Enter inserts a line"}
			case "MaxRows":
				editor.Hint = &Text{ZH: "仅 AutoResize=true 时生效", EN: "Only applies with AutoResize=true"}
			}
		}
		if t == reflect.TypeFor[nui.InputDateOptions]() || t == reflect.TypeFor[nui.InputTimeOptions]() {
			switch field.Name {
			case "Placeholder", "Password", "MaxLength":
				editor.Hint = &Text{ZH: "分段输入不适用", EN: "Not used by segments"}
			case "UI":
				editor.Slots = []string{"root", "editor", "segment", "leading", "trailing"}
				if t == reflect.TypeFor[nui.InputDateOptions]() {
					editor.Slots = append(editor.Slots, "separator")
				}
			case "Style":
				editor.Slots = []string{"root", "editor", "leading", "trailing"}
			}
		}
		if t == reflect.TypeFor[nui.InputDateOptions]() {
			switch field.Name {
			case "OnRangeChange", "SeparatorIcon", "BuildSeparator":
				editor.Hint = &Text{ZH: "仅 InputDateRange 使用", EN: "Used only by InputDateRange"}
			case "OnValueChange":
				editor.Hint = &Text{ZH: "仅 InputDate 使用；范围使用 OnRangeChange", EN: "InputDate only; ranges use OnRangeChange"}
			}
		}
		if t == reflect.TypeFor[nui.InputTimeOptions]() && field.Name == "Seconds" {
			editor.Hint = &Text{ZH: "仅 Granularity 未设置时生效", EN: "Applies only when Granularity is unset"}
		}
		if t == reflect.TypeFor[nui.PopoverOptions]() {
			switch field.Name {
			case "Label":
				editor.Hint = &Text{ZH: "无障碍名称", EN: "Accessible name"}
			case "OpenDelay", "CloseDelay":
				editor.When = &FieldCondition{Field: "Mode", Value: "nui.PopoverHover", Label: Text{ZH: "仅悬停", EN: "Hover only"}}
			case "UI", "Style":
				editor.Slots = []string{"panel"}
			}
		}
		if t == reflect.TypeFor[nui.TooltipOptions]() {
			switch field.Name {
			case "UI", "Style":
				editor.Slots = []string{"panel", "label", "trailing"}
			case "DelayDuration":
				editor.Hint = &Text{ZH: "编辑器单位为毫秒；0 表示立即显示", EN: "Editor uses milliseconds; 0 shows immediately"}
			}
		}
		if t == reflect.TypeFor[nui.ModalOptions]() || t == reflect.TypeFor[nui.SlideoverOptions]() {
			switch field.Name {
			case "UI", "Style":
				editor.Slots = []string{"backdrop", "panel", "header", "label", "description", "close", "body", "footer"}
			case "Close", "CloseIcon", "CloseLabel":
				editor.Hint = &Text{ZH: "仅 CloseButton 未关闭时生效", EN: "Applies while CloseButton is enabled"}
			case "OnClosePrevent":
				editor.When = &FieldCondition{Field: "Dismissible", Value: "false", Label: Text{ZH: "仅阻止关闭时", EN: "Blocked dismissal only"}}
			}
		}
		if t == reflect.TypeFor[nui.SlideoverOptions]() {
			switch field.Name {
			case "Width", "MaxHeight", "Fullscreen":
				editor.Hint = &Text{ZH: "继承字段，不参与 Slideover 布局；使用 Side / Extent / Inset", EN: "Inherited but unused by Slideover layout; use Side / Extent / Inset"}
			}
		}
		if t == reflect.TypeFor[nui.ToastOptions]() {
			switch field.Name {
			case "UI":
				editor.Slots = []string{"panel", "body", "leading", "label", "description", "actions", "close"}
			case "Avatar":
				editor.Hint = &Text{ZH: "优先于 Icon", EN: "Takes precedence over Icon"}
			case "CloseOptions", "CloseIcon":
				editor.Hint = &Text{ZH: "仅 Close 未关闭时使用", EN: "Used while Close is enabled"}
			case "ProgressOptions":
				editor.Hint = &Text{ZH: "仅 Progress 开启且通知不是常驻时使用", EN: "Used for enabled progress on timed notices"}
			}
		}
		if t == reflect.TypeFor[nui.ToastViewportOptions]() {
			switch field.Name {
			case "UI", "Style":
				editor.Slots = []string{"root"}
			case "Limit":
				editor.Hint = &Text{ZH: "MyGo 最多提供 3 条可见通知", EN: "MyGo supplies at most three visible notices"}
			}
		}
		if t == reflect.TypeFor[nui.ToastProgressOptions]() && field.Name == "UI" {
			editor.Slots = []string{"root", "indicator"}
		}
		if t == reflect.TypeFor[nui.MenuOptions]() || t == reflect.TypeFor[nui.MenuContentOptions]() {
			switch field.Name {
			case "UI", "Style":
				editor.Slots = []string{"panel", "header", "item", "body", "leading", "label", "description", "trailing", "indicator", "separator", "editor", "empty"}
			case "FilterFields", "IgnoreFilter", "FilterInput", "SearchTerm", "OnSearchTermChange":
				editor.Hint = &Text{ZH: "仅 Filter=true 时生效", EN: "Only applies with Filter=true"}
			}
		}
		if t == reflect.TypeFor[nui.MenuItem]() && field.Name == "UI" {
			editor.Slots = []string{"header", "item", "body", "leading", "label", "description", "trailing", "indicator", "separator"}
		}
		if t == reflect.TypeFor[nui.TabsOptions]() && (field.Name == "UI" || field.Name == "Style") {
			editor.Slots = []string{"root", "list", "trigger", "leading", "label", "trailing", "indicator", "body"}
		}
		if t == reflect.TypeFor[nui.TabItem]() && field.Name == "UI" {
			editor.Slots = []string{"trigger", "leading", "label", "trailing", "indicator", "body"}
		}
		if t == reflect.TypeFor[nui.CollapsibleOptions]() && (field.Name == "UI" || field.Name == "Style") {
			editor.Slots = []string{"root", "trigger", "leading", "label", "trailing", "body"}
		}
		if t == reflect.TypeFor[nui.AccordionOptions]() && (field.Name == "UI" || field.Name == "Style") {
			editor.Slots = []string{"root", "item", "header", "trigger", "leading", "label", "trailing", "body"}
		}
		if t == reflect.TypeFor[nui.AccordionItem]() && field.Name == "UI" {
			editor.Slots = []string{"item", "header", "trigger", "leading", "label", "trailing", "body"}
		}
		if field.Name == "UI" || field.Name == "Style" {
			if strings.HasPrefix(t.Name(), "TableColumn[") {
				editor.Slots = []string{"item", "label"}
			}
			switch t {
			case reflect.TypeFor[nui.SidebarOptions]():
				editor.Slots = []string{"root", "header", "label", "description", "actions", "trigger", "body", "footer", "separator", "panel", "backdrop", "close"}
			case reflect.TypeFor[nui.HeaderOptions]():
				editor.Slots = []string{"root", "leading", "body", "label", "description", "actions"}
			case reflect.TypeFor[nui.FooterOptions]():
				editor.Slots = []string{"root", "label", "actions"}
			case reflect.TypeFor[nui.FooterColumnsOptions]():
				editor.Slots = []string{"root", "body", "label", "list", "item", "itemLeading", "itemLabel"}
			case reflect.TypeFor[nui.MarqueeOptions]():
				editor.Slots = []string{"root", "body"}
			case reflect.TypeFor[nui.SplitterOptions]():
				editor.Slots = []string{"root", "leading", "trailing", "separator", "indicator"}
			case reflect.TypeFor[nui.CarouselOptions]():
				editor.Slots = []string{"root", "body", "item", "previous", "next", "actions", "list", "indicator", "trigger", "empty"}
			case reflect.TypeFor[nui.CarouselItem]():
				editor.Slots = []string{"item"}
			case reflect.TypeFor[nui.ChartOptions]():
				editor.Slots = []string{"root", "body", "description", "empty", "list", "item", "label", "indicator", "xAxis", "yAxis", "center"}
			case reflect.TypeFor[nui.FileUploadOptions]():
				editor.Slots = []string{"root", "body", "leading", "label", "description", "trigger", "list", "item", "itemLeading", "itemLabel", "status", "error", "itemTrailing"}
			case reflect.TypeFor[nui.DrawerOptions]():
				editor.Slots = []string{"backdrop", "panel", "container", "thumb", "indicator", "header", "label", "description", "body", "footer", "close"}
			case reflect.TypeFor[nui.NavigationOptions]():
				editor.Slots = []string{"root", "list", "label", "body", "item", "itemLeading", "itemLabel", "description", "itemTrailing", "indicator", "panel"}
			case reflect.TypeFor[nui.NavigationItem]():
				editor.Slots = []string{"list", "label", "body", "item", "itemLeading", "itemLabel", "description", "itemTrailing", "indicator", "panel"}
			case reflect.TypeFor[nui.CommandPaletteOptions]():
				editor.Slots = []string{"root", "editor", "back", "panel", "list", "label", "item", "body", "itemLeading", "itemLabel", "itemLabelPrefix", "itemLabelSuffix", "description", "itemTrailing", "indicator", "empty", "loading", "loadingIcon"}
			case reflect.TypeFor[nui.CommandAction]():
				editor.Slots = []string{"item", "body", "itemLeading", "itemLabel", "itemLabelPrefix", "itemLabelSuffix", "description", "itemTrailing", "indicator"}
			case reflect.TypeFor[nui.TreeOptions]():
				editor.Slots = []string{"root", "list", "item", "leading", "label", "trailing", "indicator", "empty"}
			case reflect.TypeFor[nui.TreeNode]():
				editor.Slots = []string{"item", "leading", "label", "trailing"}
			case reflect.TypeFor[nui.TableOptions]():
				editor.Slots = []string{"root", "caption", "body", "panel", "item", "label", "empty", "loading", "indicator"}
			case reflect.TypeFor[nui.ColorPickerOptions]():
				editor.Slots = []string{"root", "panel", "body", "item", "label"}
			case reflect.TypeFor[nui.FormOptions]():
				editor.Slots = []string{"root"}
			case reflect.TypeFor[nui.InputNumberOptions]():
				editor.Slots = []string{"root", "editor", "leading", "trailing"}
				if field.Name == "UI" {
					editor.Slots = append(editor.Slots, "increment", "decrement")
				}
			case reflect.TypeFor[nui.InputTagsOptions]():
				editor.Slots = []string{"root", "editor", "leading", "trailing"}
				if field.Name == "UI" {
					editor.Slots = append(editor.Slots, "body", "item", "label", "close", "indicator")
				}
			case reflect.TypeFor[nui.PinInputOptions]():
				editor.Slots = []string{"root", "item", "label", "separator"}
			case reflect.TypeFor[nui.InputRatingOptions]():
				editor.Slots = []string{"root", "item", "indicator", "icon", "emptyIcon"}
			case reflect.TypeFor[nui.ListboxOptions]():
				editor.Slots = []string{"root", "list", "editor", "loading", "loadingIcon", "label", "separator", "item", "leading", "body", "itemLabel", "description", "trailing", "indicator", "empty"}
			case reflect.TypeFor[nui.ListboxItemOptions]():
				editor.Slots = []string{"label", "separator", "item", "leading", "body", "itemLabel", "description", "trailing", "indicator"}
			case reflect.TypeFor[nui.SelectMenuOptions]():
				editor.Slots = []string{"root", "trigger", "leading", "label", "trailing", "close", "editor", "list", "header", "separator", "item", "itemLeading", "body", "itemLabel", "description", "itemTrailing", "indicator", "empty"}
			case reflect.TypeFor[nui.BreadcrumbOptions]():
				editor.Slots = []string{"root", "list", "item", "trigger", "leading", "label", "separator", "ellipsis"}
			case reflect.TypeFor[nui.BreadcrumbItem]():
				editor.Slots = []string{"item", "trigger", "leading", "label", "separator"}
			case reflect.TypeFor[nui.PaginationOptions]():
				editor.Slots = []string{"root", "list", "item", "first", "previous", "next", "last", "ellipsis", "leading", "label"}
			case reflect.TypeFor[nui.StepperOptions]():
				editor.Slots = []string{"root", "list", "item", "header", "trigger", "indicator", "leading", "separator", "body", "label", "description", "panel"}
			case reflect.TypeFor[nui.StepItem]():
				editor.Slots = []string{"item", "header", "trigger", "indicator", "leading", "separator", "body", "label", "description"}
			case reflect.TypeFor[nui.TimelineOptions]():
				editor.Slots = []string{"root", "item", "header", "indicator", "separator", "body", "date", "label", "description"}
			case reflect.TypeFor[nui.TimelineItem]():
				editor.Slots = []string{"item", "header", "indicator", "separator", "body", "date", "label", "description"}
			case reflect.TypeFor[nui.BannerOptions]():
				editor.Slots = []string{"root", "body", "header", "leading", "label", "description", "actions", "trailing", "close"}
			case reflect.TypeFor[nui.ErrorOptions]():
				editor.Slots = []string{"root", "leading", "status", "label", "description", "actions"}
			}
		}
		if t == reflect.TypeFor[nui.SelectMenuOptions]() {
			switch field.Name {
			case "Filter":
				editor.Hint = &Text{ZH: "不适用；用 SearchInput 控制弹层搜索框", EN: "Not used; SearchInput controls the popup search field"}
			case "FilterInput", "FocusSearch", "SearchPlaceholder":
				editor.Hint = &Text{ZH: "SearchInput 开启时生效", EN: "Used when SearchInput is enabled"}
			case "ClearIcon", "ClearOptions", "ResetModelValueOnClear", "OnClear":
				editor.Hint = &Text{ZH: "Clearable 开启且当前有选择时生效", EN: "Used when Clearable is enabled and a selection exists"}
			case "Orientation", "ItemWidth", "SelectionBehavior":
				editor.Hint = &Text{ZH: "原生列表扩展；单选时使用替换选择", EN: "Native list extension; single mode uses replacement selection"}
			}
		}
		if t == reflect.TypeFor[nui.ListboxOptions]() && field.Name == "FilterInput" {
			editor.Hint = &Text{ZH: "Filter 开启时生效；不会自动显示筛选框", EN: "Used when Filter is enabled; does not show the field by itself"}
		}
		if t == reflect.TypeFor[nui.DrawerOptions]() {
			switch field.Name {
			case "Width", "MaxHeight", "Fullscreen":
				editor.Hint = &Text{ZH: "不适用；宽度由窗口决定，高度由 SnapPoints / Snap 控制", EN: "Unused; window width and SnapPoints/Snap determine the dimensions"}
			}
		}
		if t == reflect.TypeFor[nui.ChartOptions]() {
			switch field.Name {
			case "Thickness", "DonutInnerRadius":
				editor.When = &FieldCondition{Field: "Kind", Value: "nui.ChartDonut", Label: Text{ZH: "仅圆环图", EN: "Donut only"}}
			case "Points":
				editor.When = &FieldCondition{Field: "Kind", Value: "nui.ChartLine", Label: Text{ZH: "仅折线图", EN: "Line only"}}
			case "Stack":
				editor.Hint = &Text{ZH: "仅柱状图或面积图生效", EN: "Applies to bars and areas only"}
			case "Curve":
				editor.Hint = &Text{ZH: "仅折线图或面积图生效", EN: "Applies to lines and areas only"}
			case "TooltipOptions":
				editor.Hint = &Text{ZH: "Tooltip 开启时生效；原生文本提示", EN: "Used when Tooltip is enabled; native text tooltip"}
			}
		}
		if t == reflect.TypeFor[nui.CarouselOptions]() {
			switch field.Name {
			case "Prev", "Next", "PrevIcon", "NextIcon":
				editor.When = &FieldCondition{Field: "Arrows", Value: "true", Label: Text{ZH: "显示箭头时生效", EN: "When arrows are enabled"}}
			case "Active":
				editor.Hint = &Text{ZH: "仅门控自动播放，不禁用手动操作", EN: "Gates autoplay only; manual navigation remains enabled"}
			case "StopOnInteraction":
				editor.Hint = &Text{ZH: "Autoplay 为正时生效，默认 true", EN: "Used with positive Autoplay; defaults to true"}
			}
		}
		if t == reflect.TypeFor[nui.FileUploadOptions]() {
			switch field.Name {
			case "Position":
				editor.Hint = &Text{ZH: "仅 UploadArea / UploadList 生效；网格和按钮模式保持外置列表", EN: "Only applies to UploadArea / UploadList; grid and button modes keep files outside"}
			case "FileImage":
				editor.Hint = &Text{ZH: "显示应用提供的 Avatar；组件不读取文件内容生成图片", EN: "Shows an application-provided Avatar; does not read files to create images"}
			case "Dialog":
				editor.Hint = &Text{ZH: "只控制原生选择框；Multiple 不限制队列总数，使用 Rules.MaxFiles 限制数量", EN: "Native picker only; Multiple does not limit the queue, use Rules.MaxFiles for count limits"}
			}
		}
		if t == reflect.TypeFor[nui.NavigationOptions]() {
			switch field.Name {
			case "Compact", "Single", "Collapsible", "OnExpandedChange":
				editor.When = &FieldCondition{Field: "Vertical", Value: "true", Label: Text{ZH: "仅纵向导航", EN: "Vertical navigation only"}}
			case "Popover", "PopoverOptions", "PopupPlacement":
				editor.Hint = &Text{ZH: "用于紧凑纵向导航的侧向弹层", EN: "Used by collapsed vertical navigation popovers"}
			case "Tooltip", "TooltipOptions":
				editor.Hint = &Text{ZH: "全局开关用于紧凑纵向导航；横向提示由单项 Tooltip 启用", EN: "Global toggle applies to collapsed vertical navigation; horizontal tooltips use item Tooltip"}
			}
		}
		if t == reflect.TypeFor[nui.NavigationItem]() && !slices.Contains([]string{"ID", "Label", "Icon", "Disabled", "Active", "OnSelect", "UI", "Style"}, field.Name) {
			editor.Hint = &Text{ZH: "NavigationMenu 配置；FooterColumns 当前不使用此字段", EN: "NavigationMenu configuration; currently unused by FooterColumns"}
		}
		if t == reflect.TypeFor[nui.NavigationItem]() && (field.Name == "UI" || field.Name == "Style") {
			editor.Hint = &Text{ZH: "FooterColumns 使用 Item、ItemLeadingSlot、ItemLabelSlot；其他部件用于 NavigationMenu", EN: "FooterColumns uses Item, ItemLeadingSlot and ItemLabelSlot; other parts belong to NavigationMenu"}
		}
		if t == reflect.TypeFor[nui.InputMenuOptions]() {
			switch field.Name {
			case "Multiple", "SelectedIcon", "ResetSearchTermOnSelect", "ResetModelValueOnClear":
				editor.When = &FieldCondition{Field: "Mode", Value: "nui.InputMenuCombobox", Label: Text{ZH: "仅组合框模式", EN: "Combobox only"}}
			case "DeleteIcon", "OnValuesChange":
				editor.Hint = &Text{ZH: "仅多选组合框", EN: "Multiple combobox only"}
			case "AllowCustom":
				editor.Hint = &Text{ZH: "Autocomplete 始终允许自由文本提交", EN: "Autocomplete always allows free-text submission"}
			case "OnSubmit":
				editor.Hint = &Text{ZH: "仅 AllowCustom 或 Autocomplete，且未激活候选时调用", EN: "AllowCustom or Autocomplete only, when no suggestion is activated"}
			}
		}
		fields = append(fields, editor)
	}
	return fields
}

// componentFields describes a binding in the context of its actual call. Type
// aliases retain their shared reflect.Type and configuration ID, but not every
// field of that type is meaningful to every control using it.
func componentFields(t reflect.Type, component string) []ValueField {
	fields := fieldSchemas(t)
	if component == "ContextMenu" && t == reflect.TypeFor[nui.MenuOptions]() {
		for i := range fields {
			if slices.Contains([]string{"Filter", "FilterFields", "IgnoreFilter", "SearchTerm", "FilterInput", "OnSearchTermChange"}, fields[i].Name) {
				fields[i].Hint = &Text{ZH: "原生扩展；Nuxt ContextMenu 无筛选框", EN: "Native extension; Nuxt ContextMenu has no filter input"}
			}
			switch fields[i].Name {
			case "Disabled":
				fields[i].Description = Text{ZH: "禁用上下文菜单并关闭当前面板；保留锚点自身的点击、输入和焦点能力。", EN: "Disables and closes the context menu while preserving the target’s own click, input and focus behavior."}
			case "Placement", "Offset":
				fields[i].Hint = &Text{ZH: "原生扩展；根菜单默认 RightStart、2 DIP", EN: "Native extension; root defaults to RightStart and 2 DIP"}
			}
		}
	}
	owner := ""
	switch {
	case (component == "Checkbox" || component == "CheckboxTri") && (t == reflect.TypeFor[nui.CheckboxOptions]() || t == reflect.TypeFor[nui.CheckboxTriOptions]()):
		owner = "CheckboxOptions"
	case component == "Switch" && t == reflect.TypeFor[nui.SwitchOptions]():
		owner = "SwitchOptions"
	case component == "CheckboxGroup" && t == reflect.TypeFor[nui.CheckboxGroupOptions]():
		owner = "CheckboxGroupOptions"
	case component == "RadioGroup" && t == reflect.TypeFor[nui.RadioGroupOptions]():
		owner = "RadioGroupOptions"
	}
	if owner == "" {
		return fields
	}
	notes := APINotes()
	for i := range fields {
		field := &fields[i]
		if description, ok := notes[owner+"."+field.Name]; ok {
			field.Description = description
		}
		switch {
		case owner == "CheckboxOptions" && slices.Contains([]string{"Loading", "LoadingIcon", "UncheckedIcon"}, field.Name), owner == "RadioGroupOptions" && field.Name == "Max":
			field.Hint = &Text{ZH: "不适用", EN: "Not used"}
		case owner == "CheckboxOptions" && field.Name == "Variant":
			field.Hint = &Text{ZH: "ChoiceTable 为原生扩展", EN: "ChoiceTable is a native extension"}
		case owner == "SwitchOptions" && (field.Name == "Variant" || field.Name == "Indicator"):
			field.Hint = &Text{ZH: "原生扩展", EN: "Native extension"}
		}
	}
	return fields
}

func enumNames(t reflect.Type, owner, field string) []string {
	if t == reflect.TypeFor[nui.Side]() && owner == "SidebarOptions" {
		return []string{"nui.SideLeft", "nui.SideRight"}
	}
	var values []string
	for _, v := range enumValues[t] {
		values = append(values, v.Name)
	}
	if t == reflect.TypeFor[nui.Variant]() {
		var allowed []string
		switch owner {
		case "NavigationOptions":
			allowed = []string{"nui.VariantDefault", "nui.Soft", "nui.LinkVariant"}
		case "InputOptions", "TextareaOptions", "SelectOptions", "InputMenuOptions", "InputNumberOptions", "InputTagsOptions", "PinInputOptions", "SelectMenuOptions", "InputDateOptions", "InputTimeOptions":
			allowed = []string{"nui.VariantDefault", "nui.Outline", "nui.Soft", "nui.Subtle", "nui.Ghost", "nui.None"}
		case "BadgeOptions", "AlertOptions", "CardOptions", "KbdOptions", "CalendarOptions":
			allowed = []string{"nui.VariantDefault", "nui.Solid", "nui.Outline", "nui.Soft", "nui.Subtle"}
		case "EmptyOptions":
			allowed = []string{"nui.VariantDefault", "nui.Solid", "nui.Outline", "nui.Soft", "nui.Subtle", "nui.None"}
		case "ButtonOptions", "ComponentTheme", "PaginationOptions":
			allowed = []string{"nui.VariantDefault", "nui.Solid", "nui.Outline", "nui.Soft", "nui.Subtle", "nui.Ghost", "nui.LinkVariant"}
		}
		if len(allowed) > 0 && (field == "Variant" || field == "ActiveVariant") {
			values = allowed
		}
	}
	return values
}

// Schemas uses reflection only for exported configuration values. Handles,
// bitmaps and callbacks have dedicated editors instead of arbitrary Go code.
func Schemas() map[string]ValueSchema {
	out := map[string]ValueSchema{}
	var visit func(reflect.Type)
	visit = func(t reflect.Type) {
		name := goType(t)
		if _, ok := out[name]; ok {
			return
		}
		s := ValueSchema{Type: name, Kind: "object", Sample: map[string]any{}}
		out[name] = s
		switch {
		case t == reflect.TypeFor[*mygo.Window]():
			s.Kind = "window"
			s.Sample = "current"
		case t == reflect.TypeFor[ui.Color]():
			s.Kind = "color"
			s.Sample = "#22c55e"
		case t == reflect.TypeFor[*ui.SVG]():
			s.Kind = "icon"
			s.Choices = iconNames()
			s.Sample = "check"
		case t == reflect.TypeFor[*ui.Bitmap]():
			s.Kind = "image"
			s.Sample = nil
		case t == reflect.TypeFor[*ui.Handle]():
			s.Kind = "handle"
			s.Sample = "target"
		case t == reflect.TypeFor[*ui.ScrollState]():
			s.Kind = "scroll"
			s.Sample = map[string]any{"X": float64(0), "Y": float64(0)}
		case t == reflect.TypeFor[nui.Date]():
			s.Kind = "date"
			s.Sample = "2026-11-02"
		case t == reflect.TypeFor[time.Duration]():
			s.Kind = "duration"
			s.Sample = float64(150)
		case t == reflect.TypeFor[ui.Modifiers]():
			s.Kind = "flags"
			s.Choices = []string{"ui.Ctrl", "ui.Shift", "ui.Alt", "ui.Super"}
			s.Sample = []any{}
		case optional(t):
			s.Kind = "optional"
			s.Item = goType(t.Field(0).Type)
			s.Sample = nil
			visit(t.Field(0).Type)
		case len(enumValues[t]) > 0:
			s.Kind = "enum"
			s.Choices = enumNames(t, "", "")
			s.Sample = s.Choices[0]
		default:
			switch t.Kind() {
			case reflect.String:
				s.Kind = "text"
				s.Sample = ""
			case reflect.Bool:
				s.Kind = "boolean"
				s.Sample = false
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				s.Kind = "integer"
				s.Sample = float64(0)
			case reflect.Float32, reflect.Float64:
				s.Kind = "number"
				s.Sample = float64(0)
			case reflect.Struct:
				s.Fields = fieldSchemas(t)
				for _, field := range publicFields(t) {
					visit(field.Type)
				}
			case reflect.Pointer:
				s.Kind = "pointer"
				s.Item = goType(t.Elem())
				s.Sample = nil
				visit(t.Elem())
			case reflect.Array, reflect.Slice:
				s.Kind = "array"
				s.Item = goType(t.Elem())
				s.Sample = []any{}
				if t.Kind() == reflect.Array {
					s.Length = t.Len()
				}
				visit(t.Elem())
			case reflect.Map:
				s.Kind = "map"
				s.Item = goType(t.Elem())
				for _, v := range enumValues[t.Key()] {
					s.Choices = append(s.Choices, fmt.Sprint(v.Value.Interface()))
				}
				visit(t.Elem())
			case reflect.Func:
				s.Kind = "function"
				s.Choices = functionModes(t)
				s.Sample = map[string]any{"mode": s.Choices[0], "text": "Example"}
				for i := 0; i < t.NumIn(); i++ {
					s.Arguments = append(s.Arguments, goType(t.In(i)))
				}
				for i := 0; i < t.NumOut(); i++ {
					if t.Out(i) != reflect.TypeFor[ui.Element]() {
						visit(t.Out(i))
						s.Returns = append(s.Returns, goType(t.Out(i)))
					}
				}
			case reflect.Interface:
				s.Kind = "text"
				s.Sample = ""
			default:
				panic("missing field editor for " + name)
			}
		}
		out[name] = s
	}
	for _, typ := range optionTypes {
		visit(typ)
		visit(reflect.SliceOf(typ))
	}
	return out
}

func nodeText(n ast.Node) string {
	var b bytes.Buffer
	_ = format.Node(&b, token.NewFileSet(), n)
	return b.String()
}

func expressionType(expr ast.Expr) reflect.Type {
	if nodeText(expr) == "string" {
		return reflect.TypeFor[string]()
	}
	if typ := optionTypes[nodeText(expr)]; typ != nil {
		return typ
	}
	if a, ok := expr.(*ast.ArrayType); ok && a.Len == nil {
		if typ := expressionType(a.Elt); typ != nil {
			return reflect.SliceOf(typ)
		}
	}
	return nil
}

func (d Demo) Bindings() []Binding {
	return append(d.bindingsFor(d.RenderBody(Config{}, true)), d.dataBindings()...)
}

func (d Demo) bindingsFor(body string) []Binding {
	prefix := "package main\nfunc example(){\n"
	set := token.NewFileSet()
	f, err := parser.ParseFile(set, "example.go", prefix+body+"\n}", 0)
	if err != nil {
		panic(err)
	}
	// Options can be shared with a trigger callback through a local variable.
	// Resolve only names with one definition/assignment in this example, so a
	// shadowed or reassigned name cannot patch an unrelated initializer.
	assignments := map[string]int{}
	literals := map[string]*ast.CompositeLit{}
	remember := func(name string, expr ast.Expr) {
		assignments[name]++
		literal, _ := expr.(*ast.CompositeLit)
		literals[name] = literal
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.AssignStmt:
			for i, left := range v.Lhs {
				if name, ok := left.(*ast.Ident); ok {
					var right ast.Expr
					if i < len(v.Rhs) {
						right = v.Rhs[i]
					}
					remember(name.Name, right)
				}
			}
		case *ast.ValueSpec:
			for i, name := range v.Names {
				var right ast.Expr
				if i < len(v.Values) {
					right = v.Values[i]
				}
				remember(name.Name, right)
			}
		case *ast.Field:
			for _, name := range v.Names {
				remember(name.Name, nil)
			}
		case *ast.RangeStmt:
			for _, expr := range []ast.Expr{v.Key, v.Value} {
				if name, ok := expr.(*ast.Ident); ok {
					remember(name.Name, nil)
				}
			}
		}
		return true
	})
	literalOf := func(expr ast.Expr) *ast.CompositeLit {
		if name, ok := expr.(*ast.Ident); ok && assignments[name.Name] == 1 {
			return literals[name.Name]
		}
		literal, _ := expr.(*ast.CompositeLit)
		return literal
	}
	var result []Binding
	counts := map[string]int{}
	bound := map[ast.Expr]bool{}
	add := func(expr ast.Expr, name, component string) {
		if bound[expr] {
			return
		}
		literal, ok := expr.(*ast.CompositeLit)
		if !ok {
			return
		}
		typ := expressionType(literal.Type)
		if typ == nil {
			return
		}
		bound[expr] = true
		counts[name]++
		id := name
		if counts[name] > 1 {
			id += strconv.Itoa(counts[name])
		}
		fields := componentFields(typ, component)
		for i := range fields {
			expr := literalField(literal, typ, fields[i].Name)
			key := expressionControl(expr)
			if key != "" && fields[i].Name != "AccessibleLabel" {
				fields[i].Control = key
			}
		}
		cfg, _ := d.normalizeValues(Config{})
		defaults := literalDefaults(literal, typ, cfg)
		result = append(result, Binding{ID: id, Name: name, Type: goType(typ), Fields: fields, Defaults: defaults, typ: typ, start: set.Position(expr.Pos()).Offset - len(prefix), end: set.Position(expr.End()).Offset - len(prefix), literal: expr})
	}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		callee := call.Fun
		if generic, ok := callee.(*ast.IndexExpr); ok {
			callee = generic.X
		}
		if generic, ok := callee.(*ast.IndexListExpr); ok {
			callee = generic.X
		}
		fun, ok := callee.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := fun.X.(*ast.Ident)
		if !ok || pkg.Name != "nui" || !slices.Contains(d.Symbols, fun.Sel.Name) {
			return true
		}
		for _, arg := range call.Args {
			if lit := literalOf(arg); lit != nil {
				if typ := expressionType(lit.Type); typ != nil {
					name := typ.Name()
					if name == "" {
						name = fun.Sel.Name + "Items"
					}
					add(lit, name, fun.Sel.Name)
				}
			}
		}
		return true
	})
	if d.ID == "theme" {
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok && nodeText(lit.Type) == "nui.ComponentTheme" {
				add(lit, "ComponentTheme", "")
			}
			return true
		})
		result = append(result, Binding{ID: "Theme", Name: "Theme", Type: "nui.Theme", Fields: fieldSchemas(reflect.TypeFor[nui.Theme]()), Defaults: map[string]any{}, typ: reflect.TypeFor[nui.Theme](), start: -1, end: -1})
	}
	if d.ID == "app" {
		result = append(result, Binding{ID: "AppConfig", Name: "AppConfig", Type: "nui.AppConfig", Fields: fieldSchemas(reflect.TypeFor[nui.AppConfig]()), Defaults: map[string]any{}, typ: reflect.TypeFor[nui.AppConfig](), start: -1, end: -1})
	}
	return result
}

func literalField(lit *ast.CompositeLit, typ reflect.Type, name string) ast.Expr {
	if typ.Kind() != reflect.Struct {
		return nil
	}
	for _, e := range lit.Elts {
		if kv, ok := e.(*ast.KeyValueExpr); ok {
			if key, ok := kv.Key.(*ast.Ident); ok {
				if key.Name == name {
					return kv.Value
				}
				if field, ok := typ.FieldByName(key.Name); ok && field.Anonymous {
					if nested, ok := kv.Value.(*ast.CompositeLit); ok {
						if found := literalField(nested, field.Type, name); found != nil {
							return found
						}
					}
				}
			}
		}
	}
	return nil
}

func expressionControl(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	key := ""
	count := 0
	ast.Inspect(expr, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := fn.X.(*ast.Ident)
		if !ok {
			return true
		}
		index := 0
		if pkg.Name == "showcase" && fn.Sel.Name == "Select" {
			index = 1
		} else if pkg.Name != "cfg" {
			return true
		}
		if len(call.Args) > index {
			if lit, ok := call.Args[index].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				key, _ = strconv.Unquote(lit.Value)
				count++
			}
		}
		return true
	})
	if count == 1 {
		return key
	}
	return ""
}

func literalDefaults(expr ast.Expr, typ reflect.Type, cfg Config) any {
	if expr == nil {
		return nil
	}
	if lit, ok := expr.(*ast.BasicLit); ok {
		if lit.Kind == token.STRING {
			s, _ := strconv.Unquote(lit.Value)
			return s
		}
		n, err := strconv.ParseFloat(lit.Value, 64)
		if err == nil {
			return n
		}
	}
	if id, ok := expr.(*ast.Ident); ok {
		if id.Name == "true" {
			return true
		}
		if id.Name == "false" {
			return false
		}
		return nil
	}
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		for _, entry := range enumValues[typ] {
			if entry.Name == nodeText(sel) {
				return entry.Name
			}
		}
		return nil
	}
	if optional(typ) {
		if call, ok := expr.(*ast.CallExpr); ok && nodeText(call.Fun) == "nui.Set" && len(call.Args) == 1 {
			return literalDefaults(call.Args[0], typ.Field(0).Type, cfg)
		}
	}
	if call, ok := expr.(*ast.CallExpr); ok {
		name := nodeText(call.Fun)
		if strings.HasPrefix(name, "cfg.") && len(call.Args) > 0 {
			if s, ok := literalDefaults(call.Args[0], reflect.TypeFor[string](), cfg).(string); ok {
				return cfg.Values[s]
			}
		}
		if name == "showcase.Select" && len(call.Args) == 3 {
			key, _ := literalDefaults(call.Args[1], reflect.TypeFor[string](), cfg).(string)
			if m, ok := call.Args[2].(*ast.CompositeLit); ok {
				for _, e := range m.Elts {
					if kv, ok := e.(*ast.KeyValueExpr); ok {
						if literalDefaults(kv.Key, reflect.TypeFor[string](), cfg) == cfg.Values[key] {
							return literalDefaults(kv.Value, typ, cfg)
						}
					}
				}
			}
		}
		if name == "cmp.Or" && len(call.Args) > 0 {
			for _, arg := range call.Args {
				v := literalDefaults(arg, typ, cfg)
				if v != nil && v != "" && v != false && v != float64(0) {
					return v
				}
			}
			return nil
		}
		if len(call.Args) == 1 && (name == "float32" || name == "float64" || name == "int") {
			return literalDefaults(call.Args[0], typ, cfg)
		}
		if strings.HasPrefix(name, "lucide.") {
			for _, icon := range iconNames() {
				if iconFunction(icon) == name+"()" {
					return icon
				}
			}
		}
	}
	if lit, ok := expr.(*ast.CompositeLit); ok {
		if optional(typ) {
			return nil
		}
		if typ.Kind() == reflect.Struct {
			out := map[string]any{}
			for _, field := range publicFields(typ) {
				if e := literalField(lit, typ, field.Name); e != nil {
					out[field.Name] = literalDefaults(e, field.Type, cfg)
				}
			}
			return out
		}
		if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
			out := []any{}
			for _, e := range lit.Elts {
				out = append(out, literalDefaults(e, typ.Elem(), cfg))
			}
			return out
		}
	}
	if _, ok := expr.(*ast.FuncLit); ok {
		return map[string]any{"mode": "example"}
	}
	return nil
}

func cloneJSON(value any) any {
	b, _ := json.Marshal(value)
	var out any
	_ = json.Unmarshal(b, &out)
	return out
}
