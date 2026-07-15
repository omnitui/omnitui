# OmniTUI — estrutura prevista do código-fonte

Este documento propõe a estrutura de pastas e arquivos para implementar a API definida em [API.md](API.md), a arquitetura de [DESIGN.md](DESIGN.md) e os exemplos de [COMPONENTS.md](COMPONENTS.md).

## 1. Decisão de empacotamento

O framework terá dois pacotes públicos:

```go
import (
    omnitui "github.com/viniciusfonseca/omnitui"
    components "github.com/viniciusfonseca/omnitui/components"
)
```

- `omnitui` contém runtime, elementos, componentes, estado, contexto e eventos.
- `omnitui/components` exporta `Box`, `Row`, `Column`, `Text`, `Button`, `Input`, `Tabs` e `List`.

Uma aplicação poderá usar um alias curto sem alterar o nome real do pacote:

```go
import ui "github.com/viniciusfonseca/omnitui/components"

view := ui.Row(ui.RowProps{Gap: 1}, ...)
```

O pacote raiz não importa `components`. Essa direção é obrigatória para evitar ciclo: componentes builtin dependem do runtime, mas o runtime não depende do catálogo de componentes.

## 2. Princípios

1. `omnitui` e `components` são as únicas APIs públicas do MVP.
2. `components` importa `omnitui`; o inverso é proibido.
3. Uma representação opaca em `internal/core` permite que os dois pacotes construam e consumam `Element` sem expor hosts internos.
4. Reconciliação e runtime permanecem no pacote raiz enquanto essa for a fronteira mais simples.
5. Algoritmos independentes e detalhes de plataforma ficam sob `internal/`.
6. Pacotes internos nunca importam `omnitui` nem `components`.
7. Testes ficam próximos ao código; cenários completos usam backend headless.
8. Pastas de recursos adiados não são criadas preventivamente.

## 3. Árvore prevista

```text
omnitui/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
│
├── docs/
│   ├── DESIGN.md
│   ├── API.md
│   ├── COMPONENTS.md
│   └── STRUCTURE.md
│
├── app.go
├── options.go
├── element.go
├── component.go
├── context.go
├── state.go
├── event.go
├── event_key.go
├── event_mouse.go
├── style.go
├── size.go
├── geometry.go
│
├── instance.go
├── reconcile.go
├── reconcile_children.go
├── dispatch.go
├── focus.go
├── mouse.go
├── hit_testing.go
├── runtime.go
├── paint.go
│
├── components/
│   ├── doc.go
│   ├── box.go
│   ├── row.go
│   ├── column.go
│   ├── text.go
│   ├── button.go
│   ├── input.go
│   ├── tabs.go
│   ├── list.go
│   ├── row_test.go
│   ├── column_test.go
│   ├── text_test.go
│   ├── input_test.go
│   ├── tabs_test.go
│   └── list_test.go
│
├── internal/
│   ├── core/
│   │   ├── element.go
│   │   ├── component.go
│   │   ├── host.go
│   │   ├── host_box.go
│   │   ├── host_text.go
│   │   ├── host_editable.go
│   │   ├── handler.go
│   │   ├── style.go
│   │   └── geometry.go
│   │
│   ├── backend/
│   │   ├── backend.go
│   │   ├── event.go
│   │   ├── ansi/
│   │   │   ├── backend.go
│   │   │   ├── input.go
│   │   │   ├── parser.go
│   │   │   ├── mouse.go
│   │   │   ├── parser_test.go
│   │   │   ├── mouse_test.go
│   │   │   ├── raw_unix.go
│   │   │   ├── resize_unix.go
│   │   │   └── restore_test.go
│   │   └── headless/
│   │       ├── backend.go
│   │       ├── recorder.go
│   │       └── backend_test.go
│   │
│   ├── layout/
│   │   ├── constraints.go
│   │   ├── node.go
│   │   ├── measure.go
│   │   ├── arrange.go
│   │   ├── clip.go
│   │   └── layout_test.go
│   │
│   ├── screen/
│   │   ├── cell.go
│   │   ├── buffer.go
│   │   ├── diff.go
│   │   ├── ansi.go
│   │   └── diff_test.go
│   │
│   └── text/
│       ├── grapheme.go
│       ├── width.go
│       ├── wrap.go
│       ├── truncate.go
│       └── text_test.go
│
├── examples/
│   ├── counter/
│   │   └── main.go
│   ├── form/
│   │   └── main.go
│   └── catalog/
│       └── main.go
│
├── integration/
│   ├── app_test.go
│   ├── state_test.go
│   ├── events_test.go
│   ├── mouse_test.go
│   ├── components_test.go
│   └── terminal_test.go
│
└── testdata/
    ├── screens/
    │   ├── counter_initial.txt
    │   ├── counter_incremented.txt
    │   ├── form_focused.txt
    │   └── catalog_list_scrolled.txt
    └── input/
        ├── ansi_sequences.json
        └── mouse_sequences.json
```

Essa é a estrutura-alvo do MVP. A seção 9 indica quando cada grupo deve ser criado.

## 4. Pacote público `omnitui`

### 4.1 API fundamental

A lista canônica de tipos, funções e comportamento público está em [API.md](API.md). Esta seção registra apenas a responsabilidade física dos arquivos.

| Arquivo | Responsabilidade |
|---|---|
| `app.go` | `App`, `New`, `Run`, `UpdateRoot` e `Dispatch` |
| `options.go` | `Options`, defaults e validação |
| `element.go` | `Element`, `Children`, `WithKey`, `None` e `Fragment` |
| `component.go` | `Component`, `ComponentType`, `Define` e `Create` |
| `context.go` | Contexto de render e providers tipados |
| `state.go` | `SetState`, `UpdateState` e atualizações pendentes |
| `event.go` | Eventos, handlers, `Propagate` e `Consume` |
| `event_key.go` | Teclas, runes e modificadores públicos |
| `event_mouse.go` | Ações, botões, coordenadas, `MouseEvent` e `WheelEvent` |
| `style.go` | Cores, atributos e `Style` |
| `size.go` | `Size`, `Cells` e dimensões automáticas |
| `geometry.go` | `Spacing`, `Rect` e helpers como `All` |

O pacote raiz não exporta componentes visuais. Em especial, não existem `omnitui.Text`, `omnitui.Row` ou `omnitui.List`; esses símbolos pertencem a `components`.

### 4.2 Runtime e reconciliação

| Arquivo | Responsabilidade |
|---|---|
| `instance.go` | Instâncias montadas, identidade, estado, props e caminho de diagnóstico |
| `reconcile.go` | Mount, update, replace e unmount |
| `reconcile_children.go` | Reconciliação posicional e por chave |
| `dispatch.go` | Fila serializada de estado, mensagens, resize e input |
| `focus.go` | Ordem de foco, alvos e recuperação após unmount |
| `mouse.go` | Hover path, captura, estado dos botões e comportamentos padrão |
| `hit_testing.go` | Busca do alvo por posição, clipping e ordem de pintura |
| `runtime.go` | Loop principal e coordenação das fases |
| `paint.go` | Conversão da árvore host posicionada para o buffer traseiro |

Esses arquivos permanecem no pacote raiz e mantêm seus símbolos internos não exportados.

## 5. Pacote público `components`

### 5.1 API exportada

| Arquivo | Símbolos principais |
|---|---|
| `doc.go` | Visão geral e exemplo de importação do pacote |
| `box.go` | `Box`, `BoxProps` |
| `row.go` | `Row`, `RowProps` |
| `column.go` | `Column`, `ColumnProps` |
| `text.go` | `Text`, `TextProps`, wrapping e truncamento |
| `button.go` | `Button`, `ButtonProps` |
| `input.go` | `Input`, `InputProps` |
| `tabs.go` | `Tabs`, `TabsProps`, `TabItem` |
| `list.go` | `List`, `ListProps`, `ScrollbarMode` |

Assinaturas, props e enums estão centralizados em [API.md](API.md). Esta seção registra apenas a localização de cada implementação.

### 5.2 Implementação

- `Row`, `Column`, `Button`, `Input`, `Tabs` e `List` usam o contrato `omnitui.Component`.
- `Box` e `Text` criam hosts por meio de `internal/core`.
- `Input` usa o host editável privado de `internal/core`.
- Props públicas usam tipos fundamentais de `omnitui`, como `Style`, `Size`, `Spacing` e eventos.
- Nenhuma API de `components` é necessária para iniciar ou executar uma aplicação.

O pacote não acessa instâncias, reconciliador, filas ou backend. Ele descreve elementos e reage a eventos como qualquer biblioteca de componentes de usuário.

## 6. Fronteira compartilhada `internal/core`

Separar os builtins cria um problema específico de Go: campos não exportados do `omnitui.Element` não seriam acessíveis por `components`, enquanto exportar constructors de hosts poluiria a API do runtime.

`internal/core` resolve isso:

- define a representação opaca concreta de `Element`;
- define kinds e payloads dos hosts `Box`, `Text` e editável;
- armazena handlers apagados por tipo;
- contém valores neutros de estilo e geometria compartilhados;
- não conhece runtime, estado, contexto ou componentes builtin.

O pacote raiz reexporta aliases públicos não genéricos para a representação opaca de elemento e para valores compartilhados de estilo e geometria. Os nomes e contratos desses aliases ficam exclusivamente em [API.md](API.md).

Como `internal/core` está sob a árvore do módulo, `components` pode importá-lo, mas consumidores externos não podem. Assim, o catálogo cria hosts sem tornar essa construção parte da API pública.

`ComponentType[P]`, `Context` e helpers genéricos continuam definidos em `omnitui`; eles não precisam ser aliases de tipos internos.

## 7. Outros pacotes internos

### 7.1 `internal/backend`

Define o contrato neutro entre runtime e plataformas:

```go
type Backend interface {
    Size() (width, height int, err error)
    Events() <-chan Event
    Write([]byte) error
    Close() error
}
```

- `backend/ansi` implementa modo raw, alternate screen, input ANSI, SGR mouse, wheel, resize e restauração Unix.
- `backend/headless` recebe eventos sintéticos e captura frames para testes.
- backends não conhecem `omnitui`, `components`, foco ou estado.

### 7.2 `internal/layout`

Executa `measure` e `arrange` sobre uma árvore normalizada. Não conhece componentes, eventos ou ANSI.

### 7.3 `internal/screen`

É dono de células, buffers, clipping, diff e codificação ANSI determinística. `Style` é compilado para atributos internos antes do paint.

### 7.4 `internal/text`

Centraliza graphemes, largura visual, wrapping, truncamento e movimentação de cursor. Somente esse pacote depende diretamente da biblioteca Unicode escolhida.

## 8. Direção das dependências

```text
examples/ -------------------------------> omnitui
examples/ -------------------------------> components
integration/ ----------------------------> omnitui
integration/ ----------------------------> components
integration/ -----> internal/backend/headless

components -----> omnitui
components -----> internal/core

omnitui -----> internal/core
omnitui -----> internal/backend
omnitui -----> internal/backend/ansi
omnitui -----> internal/layout
omnitui -----> internal/screen
omnitui -----> internal/text

internal/backend/ansi -----> internal/backend
internal/backend/headless --> internal/backend

omnitui -X-> components
internal/* -X-> omnitui
internal/* -X-> components
```

Regras obrigatórias:

- `components` pode importar `omnitui`; `omnitui` nunca importa `components`;
- pacotes internos não importam nenhum pacote público do módulo;
- `internal/core` é uma estrutura de dados compartilhada, não um segundo runtime;
- `ansi` e `headless` dependem somente do contrato `backend`;
- exemplos importam somente pacotes públicos;
- testes de integração podem importar o backend headless;
- não criar pacotes genéricos `util`, `common`, `shared` ou `helpers`.

## 9. Criação incremental

### Fases 0 e 1 — contrato e reconciliação

```text
go.mod
element.go
component.go
context.go
state.go
event.go
app.go
instance.go
reconcile.go
reconcile_children.go
runtime.go
internal/core/element.go
internal/core/component.go
internal/backend/backend.go
internal/backend/headless/backend.go
examples/counter/main.go
```

### Fase 2 — hosts, layout e componentes básicos

```text
style.go
size.go
geometry.go
internal/core/host*.go
internal/core/style.go
internal/core/geometry.go
internal/layout/
internal/screen/
internal/text/
components/box.go
components/row.go
components/column.go
components/text.go
testdata/screens/
```

### Fases 3 e 4 — terminal e interação

```text
event_key.go
event_mouse.go
dispatch.go
focus.go
mouse.go
hit_testing.go
paint.go
components/button.go
internal/backend/ansi/
testdata/input/mouse_sequences.json
integration/events_test.go
integration/mouse_test.go
integration/terminal_test.go
```

### Fase 5 — componentes com estado

```text
internal/core/host_editable.go
components/input.go
components/tabs.go
components/list.go
examples/form/
examples/catalog/
integration/components_test.go
testdata/screens/catalog_list_scrolled.txt
```

### Fase 6 — endurecimento

Adicionar benchmarks, fuzz tests e novos arquivos somente onde medições indicarem necessidade. Memoização, virtualização e scheduler não ganham pastas preventivas.

## 10. Testes

- testes do runtime ficam junto ao pacote raiz;
- cada builtin tem testes em `components/*_test.go`;
- testes de algoritmos ficam sob o pacote interno correspondente;
- `integration/` usa `omnitui` e `components` como um consumidor externo;
- `integration/mouse_test.go` cobre hit testing, clipping, bubbling, captura, hover, press e wheel;
- snapshots ficam em `testdata/screens`;
- fuzzing do parser, inclusive sequências SGR mouse, fica em `internal/backend/ansi/parser_fuzz_test.go`.

Os testes de `components` devem preferir o contrato público e o backend headless. Acesso direto a `internal/core` só é aceitável para testar tradução de props em hosts.

## 11. Estruturas deliberadamente adiadas

```text
internal/backend/windows/  # suporte nativo a Windows
internal/virtual/          # listas virtualizadas
internal/animation/        # clock, scheduler e transições
internal/effects/          # lifecycle e cleanup
```

Esses nomes são marcadores conceituais, não diretórios reservados.

## 12. Critérios de aceitação

1. Aplicações importam `omnitui` para runtime e `components` para UI builtin.
2. Nenhum componente visual é exportado pelo pacote raiz.
3. `components` depende de `omnitui`, nunca o inverso.
4. Nenhum ciclo de importação é necessário.
5. Hosts internos não são expostos para consumidores externos.
6. Backends ANSI e headless implementam o mesmo contrato.
7. Reconciliador é testável sem componentes builtin ou terminal.
8. Builtins passam pelo mesmo modelo de instâncias e estado dos componentes de usuário.
9. Mouse SGR é isolado no backend; hit testing e captura permanecem no runtime.
10. `go test ./...` cobre os dois pacotes públicos e testes de integração.
11. `go test -race ./...` não encontra mutação fora da goroutine do runtime.
