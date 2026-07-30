package components

import (
	"context"
	"fmt"

	"github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/internal/core"
)

type DropdownOption struct {
	Key      string
	Label    string
	Disabled bool
}

type DropdownProps struct {
	Options             []DropdownOption
	SelectedKey         string
	Placeholder         string
	Width               omnitui.Size
	MenuHeight          omnitui.Size
	Disabled            bool
	Wrap                bool
	Style               omnitui.Style
	FocusStyle          omnitui.Style
	DisabledStyle       omnitui.Style
	MenuStyle           omnitui.Style
	SelectedStyle       omnitui.Style
	DisabledOptionStyle omnitui.Style
	OnChange            omnitui.EventHandler[omnitui.ValueChangeEvent]
}

type dropdownState struct {
	Open      bool
	ActiveKey string
}

type dropdownComponent struct{}

func (dropdownComponent) InitialState(DropdownProps) dropdownState { return dropdownState{} }

func (dropdownComponent) Render(ctx omnitui.Context, props DropdownProps, state dropdownState, _ omnitui.Children) omnitui.Element {
	triggerFocus := omnitui.UseFocus(ctx, "trigger")
	open := state.Open && !props.Disabled && firstEnabledDropdownOption(props.Options) != ""

	if state.Open && !open {
		omnitui.UseEffect(ctx, "close-unavailable", struct{}{}, func(context.Context) omnitui.Cleanup {
			omnitui.UpdateState(ctx, func(current dropdownState) dropdownState {
				current.Open = false
				current.ActiveKey = ""
				return current
			})
			return nil
		})
	}

	closeMenu := func(returnFocus bool) {
		omnitui.UpdateState(ctx, func(current dropdownState) dropdownState {
			current.Open = false
			current.ActiveKey = ""
			return current
		})
		if returnFocus {
			triggerFocus.Request()
		}
	}
	openMenu := func(preferLast bool) {
		active := props.SelectedKey
		if active == "" {
			if preferLast {
				active = lastEnabledDropdownOption(props.Options)
			} else {
				active = firstEnabledDropdownOption(props.Options)
			}
		}
		omnitui.UpdateState(ctx, func(current dropdownState) dropdownState {
			current.Open = active != ""
			current.ActiveKey = active
			return current
		})
	}
	selectOption := func(key string) {
		option, available := dropdownOption(props.Options, key)
		if !available || option.Disabled {
			return
		}
		if key != props.SelectedKey && props.OnChange != nil {
			props.OnChange(omnitui.ValueChangeEvent{
				Previous: props.SelectedKey,
				Value:    key,
				Source:   omnitui.ChangeKeyboard,
			})
		}
		closeMenu(true)
	}

	label := props.Placeholder
	if selected, ok := dropdownOption(props.Options, props.SelectedKey); ok {
		label = selected.Label
	}
	if open {
		label += " ▴"
	} else {
		label += " ▾"
	}

	children := []omnitui.Element{
		Button(ButtonProps{
			Label:         label,
			Plain:         true,
			Disabled:      props.Disabled || firstEnabledDropdownOption(props.Options) == "",
			Style:         props.Style,
			FocusStyle:    props.FocusStyle,
			DisabledStyle: props.DisabledStyle,
			Focus:         triggerFocus,
			OnKey: func(event omnitui.KeyEvent) omnitui.EventResult {
				if state.Open {
					return omnitui.Propagate
				}
				switch event.Key {
				case omnitui.KeyDown:
					openMenu(false)
					return omnitui.Consume
				case omnitui.KeyUp:
					openMenu(true)
					return omnitui.Consume
				default:
					return omnitui.Propagate
				}
			},
			OnPress: func(omnitui.PressEvent) omnitui.EventResult {
				if open {
					closeMenu(true)
				} else {
					openMenu(false)
				}
				return omnitui.Consume
			},
		}),
	}

	if open {
		menuFocus := omnitui.UseFocus(ctx, "menu")
		omnitui.UseEffect(ctx, "focus-menu", struct{}{}, func(context.Context) omnitui.Cleanup {
			menuFocus.Request()
			return nil
		})

		items := make([]omnitui.Element, 0, len(props.Options))
		for _, option := range props.Options {
			option := option
			itemStyle := omnitui.Style{}
			var onMouse omnitui.EventHandler[omnitui.MouseEvent]
			if option.Disabled {
				itemStyle = props.DisabledOptionStyle
				onMouse = func(event omnitui.MouseEvent) omnitui.EventResult {
					if event.Action == omnitui.MouseDown && event.Button == omnitui.MouseButtonLeft {
						return omnitui.Consume
					}
					return omnitui.Propagate
				}
			}
			items = append(items, Box(
				BoxProps{
					Padding:  omnitui.XY(1, 0),
					Disabled: option.Disabled,
					Style:    itemStyle,
					OnMouse:  onMouse,
					OnPress: func(omnitui.PressEvent) omnitui.EventResult {
						selectOption(option.Key)
						return omnitui.Consume
					},
				},
				Text(TextProps{Content: option.Label}),
			).WithKey(option.Key))
		}

		menu := List(
			ListProps{
				SelectedKey:   state.ActiveKey,
				Selectable:    true,
				Height:        props.MenuHeight,
				Disabled:      props.Disabled,
				Wrap:          props.Wrap,
				Scrollbar:     ScrollbarAuto,
				Style:         props.MenuStyle,
				SelectedStyle: props.SelectedStyle,
				Focus:         menuFocus,
				OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
					next := nextEnabledDropdownOption(props.Options, event.Previous, event.Value, props.Wrap)
					if next != "" && next != state.ActiveKey {
						omnitui.UpdateState(ctx, func(current dropdownState) dropdownState {
							current.ActiveKey = next
							return current
						})
					}
					return omnitui.Consume
				},
				OnActivate: func(event omnitui.ActivateEvent) omnitui.EventResult {
					selectOption(event.Key)
					return omnitui.Consume
				},
			},
			items...,
		)
		children = append(children, core.NewHost(core.HostOverlay, core.OverlayData{}, []omnitui.Element{menu}))
	}

	return Box(
		BoxProps{
			Width:     props.Width,
			Direction: Vertical,
			Align:     AlignStretch,
			OnKey: func(event omnitui.KeyEvent) omnitui.EventResult {
				if !open {
					return omnitui.Propagate
				}
				switch event.Key {
				case omnitui.KeyEscape:
					closeMenu(true)
					return omnitui.Consume
				case omnitui.KeyHome:
					omnitui.UpdateState(ctx, func(current dropdownState) dropdownState {
						current.ActiveKey = firstEnabledDropdownOption(props.Options)
						return current
					})
					return omnitui.Consume
				case omnitui.KeyEnd:
					omnitui.UpdateState(ctx, func(current dropdownState) dropdownState {
						current.ActiveKey = lastEnabledDropdownOption(props.Options)
						return current
					})
					return omnitui.Consume
				case omnitui.KeyTab, omnitui.KeyBacktab:
					closeMenu(false)
					return omnitui.Propagate
				default:
					return omnitui.Propagate
				}
			},
		},
		children...,
	)
}

var dropdownType = omnitui.Define[DropdownProps, dropdownState]("Dropdown", dropdownComponent{})

func Dropdown(props DropdownProps) omnitui.Element {
	props.Options = append([]DropdownOption(nil), props.Options...)
	if props.Placeholder == "" {
		props.Placeholder = "Select an option"
	}
	validateStyle(props.Style)
	validateStyle(props.FocusStyle)
	validateStyle(props.DisabledStyle)
	validateStyle(props.MenuStyle)
	validateStyle(props.SelectedStyle)
	validateStyle(props.DisabledOptionStyle)

	seen := make(map[string]struct{}, len(props.Options))
	for index, option := range props.Options {
		if option.Key == "" {
			panic(fmt.Sprintf("omnitui/components: empty dropdown option key at index %d", index))
		}
		if _, exists := seen[option.Key]; exists {
			panic(fmt.Sprintf("omnitui/components: duplicate dropdown option key %q", option.Key))
		}
		seen[option.Key] = struct{}{}
	}
	if props.SelectedKey != "" {
		option, ok := dropdownOption(props.Options, props.SelectedKey)
		if !ok {
			panic(fmt.Sprintf("omnitui/components: unknown selected dropdown option %q", props.SelectedKey))
		}
		if option.Disabled {
			panic(fmt.Sprintf("omnitui/components: selected dropdown option %q is disabled", props.SelectedKey))
		}
	}
	return omnitui.Create(dropdownType, props)
}

func dropdownOption(options []DropdownOption, key string) (DropdownOption, bool) {
	for _, option := range options {
		if option.Key == key {
			return option, true
		}
	}
	return DropdownOption{}, false
}

func firstEnabledDropdownOption(options []DropdownOption) string {
	for _, option := range options {
		if !option.Disabled {
			return option.Key
		}
	}
	return ""
}

func lastEnabledDropdownOption(options []DropdownOption) string {
	for index := len(options) - 1; index >= 0; index-- {
		if !options[index].Disabled {
			return options[index].Key
		}
	}
	return ""
}

func nextEnabledDropdownOption(options []DropdownOption, previous, proposed string, wrap bool) string {
	proposedIndex := dropdownOptionIndex(options, proposed)
	if proposedIndex < 0 {
		return ""
	}
	if !options[proposedIndex].Disabled {
		return proposed
	}

	previousIndex := dropdownOptionIndex(options, previous)
	direction := 1
	if previousIndex >= 0 && proposedIndex < previousIndex {
		direction = -1
	}
	if wrap && previousIndex == 0 && proposedIndex == len(options)-1 {
		direction = -1
	}
	if wrap && previousIndex == len(options)-1 && proposedIndex == 0 {
		direction = 1
	}

	index := proposedIndex
	for visited := 0; visited < len(options); visited++ {
		index += direction
		if index < 0 || index >= len(options) {
			if !wrap {
				return previous
			}
			if index < 0 {
				index = len(options) - 1
			} else {
				index = 0
			}
		}
		if !options[index].Disabled {
			return options[index].Key
		}
	}
	return previous
}

func dropdownOptionIndex(options []DropdownOption, key string) int {
	for index, option := range options {
		if option.Key == key {
			return index
		}
	}
	return -1
}
