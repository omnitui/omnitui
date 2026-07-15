# OmniTUI

OmniTUI is a declarative terminal UI framework in Go. The public packages are
`github.com/viniciusfonseca/omnitui` and `github.com/viniciusfonseca/omnitui/components`.

```go
view := components.Column(components.ColumnProps{Gap: 1},
	components.Text(components.TextProps{Content: "Olá"}),
)
app := omnitui.New(view, omnitui.Options{})
_ = app.Run(context.Background())
```

See [`docs/API.md`](docs/API.md) for the canonical API and the other documents
for the design, builtin behavior and source organization.
