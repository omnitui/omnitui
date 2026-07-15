package core

// Kind identifies the internal representation of an element.
type Kind uint8

const (
	KindNone Kind = iota
	KindFragment
	KindHost
	KindComponent
	KindProvider
)

// Element is the immutable, opaque description shared by the public packages.
type Element struct {
	kind      Kind
	key       string
	children  []Element
	host      Host
	component *ComponentDefinition
	provider  Provider
}

func None() Element { return Element{} }

func Fragment(children ...Element) Element {
	return Element{kind: KindFragment, children: cloneElements(children)}
}

func Component(def *ComponentDefinition, props any, children []Element) Element {
	return Element{kind: KindComponent, component: def, host: Host{Data: props}, children: cloneElements(children)}
}

func ProviderElement(keyID uint64, value any, child Element) Element {
	return Element{kind: KindProvider, provider: Provider{KeyID: keyID, Value: value}, children: []Element{child}}
}

func (e Element) WithKey(key string) Element {
	e.key = key
	return e
}

func KindOf(e Element) Kind  { return e.kind }
func KeyOf(e Element) string { return e.key }

func PropsOf(e Element) any { return e.host.Data }

func ChildrenOf(e Element) []Element {
	return cloneElements(e.children)
}

func ComponentOf(e Element) *ComponentDefinition { return e.component }
func ProviderOf(e Element) Provider              { return e.provider }

func cloneElements(in []Element) []Element {
	if len(in) == 0 {
		return nil
	}
	out := make([]Element, len(in))
	copy(out, in)
	return out
}
