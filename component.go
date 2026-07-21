package omnitui

import "github.com/omnitui/omnitui/internal/core"

type Component[P, S any] interface {
	InitialState(props P) S
	Render(ctx Context, props P, state S, children Children) Element
}

type ComponentType[P any] struct{ definition *core.ComponentDefinition }

func Define[P, S any](name string, component Component[P, S]) ComponentType[P] {
	if component == nil {
		panic("omnitui: cannot define a nil component")
	}
	return ComponentType[P]{definition: &core.ComponentDefinition{
		Name: name,
		InitialState: func(props any) any {
			return component.InitialState(props.(P))
		},
		Render: func(ctx any, props any, state any, children []core.Element) core.Element {
			return component.Render(ctx.(Context), props.(P), state.(S), Children(children))
		},
	}}
}

func Create[P any](component ComponentType[P], props P, children ...Element) Element {
	if component.definition == nil {
		panic("omnitui: cannot create an undefined component")
	}
	return core.Component(component.definition, props, children)
}
