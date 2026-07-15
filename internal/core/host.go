package core

type HostKind uint8

const (
	HostBox HostKind = iota
	HostText
	HostButton
	HostInput
	HostTabs
	HostList
)

// Host is the normalized data consumed by reconciliation, layout and paint.
type Host struct {
	Kind     HostKind
	Data     any
	Children []Element
}

func NewHost(kind HostKind, data any, children []Element) Element {
	return Element{
		kind:     KindHost,
		host:     Host{Kind: kind, Data: cloneHostData(data), Children: cloneElements(children)},
		children: cloneElements(children),
	}
}

func HostOf(e Element) (Host, bool) {
	if e.kind != KindHost {
		return Host{}, false
	}
	h := e.host
	h.Children = cloneElements(h.Children)
	return h, true
}

func cloneHostData(data any) any {
	switch value := data.(type) {
	case BoxData:
		value.Handlers = cloneHandlers(value.Handlers)
		return value
	case TextData:
		return value
	case ButtonData:
		value.Handlers = cloneHandlers(value.Handlers)
		return value
	case InputData:
		value.Handlers = cloneHandlers(value.Handlers)
		return value
	case TabsData:
		value.Items = append([]TabData(nil), value.Items...)
		value.Handlers = cloneHandlers(value.Handlers)
		return value
	case ListData:
		value.Handlers = cloneHandlers(value.Handlers)
		return value
	default:
		return data
	}
}

type Handlers map[string]any

func cloneHandlers(in Handlers) Handlers {
	if len(in) == 0 {
		return nil
	}
	out := make(Handlers, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

type BoxData struct {
	Width, Height        Size
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
	Padding              Spacing
	Gap                  int
	Direction            uint8
	Align                uint8
	Justify              uint8
	Wrap, Clip           bool
	Border               uint8
	Style                Style
	Focusable, Disabled  bool
	Handlers             Handlers
}

type TextData struct {
	Content  string
	Style    Style
	Wrap     uint8
	Align    uint8
	MaxLines int
	Truncate uint8
}

type ButtonData struct {
	Label                            string
	Disabled                         bool
	Style, FocusStyle, DisabledStyle Style
	Handlers                         Handlers
}

type InputData struct {
	Value, Placeholder string
	Width              Size
	Disabled, ReadOnly bool
	Mask               rune
	MaxLength          int
	Style, FocusStyle  Style
	Handlers           Handlers
}

type TabData struct {
	Key      string
	Label    string
	Content  Element
	Disabled bool
}

type TabsData struct {
	Items              []TabData
	ActiveKey          string
	Orientation        uint8
	Style, ActiveStyle Style
	Handlers           Handlers
}

type ListData struct {
	SelectedKey          string
	Height               Size
	Gap                  int
	Disabled, Wrap       bool
	ScrollPadding        int
	Scrollbar            uint8
	Empty                Element
	Style, SelectedStyle Style
	Handlers             Handlers
}
