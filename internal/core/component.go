package core

// ComponentDefinition is the erased adapter used by the generic public API.
// It deliberately knows nothing about the runtime package.
type ComponentDefinition struct {
	Name         string
	InitialState func(props any) any
	Render       func(ctx any, props any, state any, children []Element) Element
}

type Provider struct {
	KeyID uint64
	Value any
}
