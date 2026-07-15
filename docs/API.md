# OmniTUI — referência da API pública

Este documento é a fonte canônica da API pública do framework. A arquitetura interna está em [DESIGN.md](DESIGN.md), os exemplos completos dos builtins em [COMPONENTS.md](COMPONENTS.md) e a organização do código em [STRUCTURE.md](STRUCTURE.md).

## 1. Pacotes públicos

```go
import (
    omnitui "github.com/viniciusfonseca/omnitui"
    components "github.com/viniciusfonseca/omnitui/components"
)
```

- `omnitui`: elementos, componentes, estado, contexto, runtime, eventos, geometria e estilos.
- `omnitui/components`: `Box`, `Row`, `Column`, `Text`, `Button`, `Input`, `Tabs` e `List`.

O caminho `github.com/viniciusfonseca/omnitui` é provisório até a definição do módulo em `go.mod`.

## 2. Elementos e componentes — `omnitui`

### `Element` e filhos

```go
type Element struct {
    // representação opaca
}

type Children []Element

func (e Element) WithKey(key string) Element
func None() Element
func Fragment(children ...Element) Element
```

- `Element` é uma descrição imutável e barata da interface.
- `WithKey` retorna uma cópia; a chave deve ser única somente entre irmãos.
- `None` representa ausência de conteúdo.
- `Fragment` agrupa vários filhos sem criar uma caixa de layout.
- O valor zero de `Element` é equivalente a `None`.

### Definição de componente

```go
type Component[P, S any] interface {
    InitialState(props P) S
    Render(ctx Context, props P, state S, children Children) Element
}

type ComponentType[P any] struct {
    // identidade opaca e estável
}

func Define[P, S any](name string, component Component[P, S]) ComponentType[P]

func Create[P any](
    component ComponentType[P],
    props P,
    children ...Element,
) Element
```

Regras:

- o valor passado a `Define` não deve guardar estado mutável;
- cada ocorrência montada de `ComponentType` possui estado independente;
- props e filhos são substituídos pelo render mais recente;
- `Render` deve ser determinístico para as mesmas entradas e não pode atualizar estado;
- múltiplos filhos de saída devem ser envolvidos em `Fragment` ou em um container;
- tipo, posição e chave determinam se uma instância será preservada.

## 3. Estado — `omnitui`

```go
func SetState[S any](ctx Context, next S)

func UpdateState[S any](
    ctx Context,
    update func(current S) S,
)
```

- `SetState` substitui o estado atual.
- `UpdateState` é preferido quando o próximo valor depende do anterior.
- Atualizações são enfileiradas, aplicadas em ordem e agrupadas em um único frame quando possível.
- É seguro chamar essas funções de outra goroutine; somente o runtime aplica a mutação.
- Atualizações para uma instância desmontada são ignoradas.
- Tipo incorreto ou atualização durante `Render` é erro de programação.

## 4. Contexto — `omnitui`

```go
type Context struct {
    // instância atual, dispatcher e valores herdados
}

type ContextKey[T any] struct {
    // identidade e valor padrão
}

func NewContext[T any](defaultValue T) ContextKey[T]
func UseContext[T any](ctx Context, key ContextKey[T]) T
func Provide[T any](key ContextKey[T], value T, child Element) Element
```

`Context` é o contexto de render do framework e não substitui `context.Context`. Providers usam escopo de árvore: o valor mais próximo vence e não vaza para irmãos.

## 5. Runtime — `omnitui`

```go
type ColorProfile uint8

const (
    ColorProfileAuto ColorProfile = iota
    ColorProfileANSI16
    ColorProfileANSI256
    ColorProfileTrueColor
)

type Options struct {
    Input        io.Reader
    Output       io.Writer
    ColorProfile ColorProfile
}

type App struct {
    // runtime opaco
}

func New(root Element, options Options) *App
func (app *App) Run(ctx context.Context) error
func (app *App) UpdateRoot(root Element)
func (app *App) Dispatch(message any)
```

- `Input` e `Output` usam o terminal atual quando omitidos.
- `ColorProfileAuto` detecta a melhor capacidade disponível; uma opção explícita torna testes e ambientes remotos previsíveis.
- `Run` assume posse do terminal até retornar e restaura seu estado em sucesso, cancelamento, erro ou panic.
- `UpdateRoot` e `Dispatch` apenas publicam trabalho na fila e podem ser chamados por outras goroutines.
- `Dispatch(value)` gera `MessageEvent` no nó host raiz.

Erro público do MVP:

```go
var ErrInterrupted = errors.New("omnitui: interrupted")
```

`Run` retorna `ErrInterrupted` quando `Ctrl+C` não é consumido.

## 6. Geometria — `omnitui`

```go
type Size struct {
    // modo e valor opacos
}

func Auto() Size
func Cells(value int) Size

type Spacing struct {
    Top, Right, Bottom, Left int
}

func All(value int) Spacing
func XY(horizontal, vertical int) Spacing

type Rect struct {
    X, Y, Width, Height int
}
```

- O valor zero de `Size` equivale a `Auto()`.
- Tamanhos e espaçamentos negativos são inválidos.
- Percentuais e unidades flexíveis não fazem parte do MVP.

## 7. Estilos — `omnitui`

### Cores

```go
type Color struct {
    // kind e canais opacos
}

type ANSIColor uint8

const (
    Black ANSIColor = iota
    Red
    Green
    Yellow
    Blue
    Magenta
    Cyan
    White
    BrightBlack
    BrightRed
    BrightGreen
    BrightYellow
    BrightBlue
    BrightMagenta
    BrightCyan
    BrightWhite
)

func DefaultColor() Color
func ANSI(color ANSIColor) Color
func Indexed(index uint8) Color
func RGB(red, green, blue uint8) Color
```

O valor zero de `Color` significa **não especificado** e herda a cor resolvida do pai. `DefaultColor()` é diferente: ele solicita explicitamente a cor padrão do terminal e interrompe a herança.

Perfis suportados:

| Perfil | Cores |
|---|---:|
| `ColorProfileANSI16` | 16 cores ANSI |
| `ColorProfileANSI256` | paleta indexada de 256 cores |
| `ColorProfileTrueColor` | RGB de 24 bits |

Quando uma cor excede o perfil ativo, o renderer escolhe a entrada visualmente mais próxima da paleta disponível. Alpha, gradientes e mistura de cores não são suportados.

### Atributos de texto

```go
type AttributeMask uint16

const (
    Bold AttributeMask = 1 << iota
    Dim
    Italic
    Underline
    Blink
    Reverse
    Hidden
    Strikethrough
)

type Style struct {
    Foreground      Color
    Background      Color
    Attributes      AttributeMask
    ClearAttributes AttributeMask
}
```

| Atributo | Efeito esperado | Observação |
|---|---|---|
| `Bold` | Intensidade forte | Pode selecionar cor brilhante em terminais antigos |
| `Dim` | Intensidade reduzida | Nem todo terminal distingue de cor normal |
| `Italic` | Texto inclinado | Pode ser ignorado pelo terminal |
| `Underline` | Sublinhado simples | Suporte amplo |
| `Blink` | Texto piscante | Frequentemente desabilitado pelo terminal |
| `Reverse` | Troca foreground e background | Suporte amplo |
| `Hidden` | Oculta o conteúdo visual | As células continuam ocupando layout |
| `Strikethrough` | Texto riscado | Pode ser ignorado pelo terminal |

O framework emite os códigos SGR correspondentes, mas não simula visualmente atributos que o terminal ignora. `DoubleUnderline`, `Overline`, hyperlinks, fontes e velocidade de blink ficam fora do MVP.

### Herança e composição

Para resolver o estilo de um nó:

1. começar pelo estilo resolvido do pai;
2. substituir foreground/background quando a cor não for zero;
3. remover os bits de `ClearAttributes`;
4. adicionar os bits de `Attributes`.

O mesmo bit não pode aparecer em `Attributes` e `ClearAttributes`; isso gera erro de props. O valor zero de `Style` herda integralmente o estilo do pai.

```go
titleStyle := omnitui.Style{
    Foreground: omnitui.ANSI(omnitui.Cyan),
    Attributes: omnitui.Bold | omnitui.Underline,
}

normalChild := omnitui.Style{
    ClearAttributes: omnitui.Bold,
}
```

Props como `FocusStyle`, `ActiveStyle` e `SelectedStyle` são aplicadas depois de `Style`, usando as mesmas regras. Padding, gap, border, alinhamento e clipping são propriedades de componente, não atributos de `Style`.

## 8. Componentes builtin — `components`

### Enums compartilhados

```go
type Direction uint8
const (
    Horizontal Direction = iota
    Vertical
)

type Align uint8
const (
    AlignStart Align = iota
    AlignCenter
    AlignEnd
    AlignStretch
)

type Justify uint8
const (
    JustifyStart Justify = iota
    JustifyCenter
    JustifyEnd
    JustifySpaceBetween
    JustifySpaceAround
)

type TextWrap uint8
const (
    WrapNone TextWrap = iota
    WrapWord
    WrapGrapheme
)

type TextAlign uint8
const (
    TextAlignStart TextAlign = iota
    TextAlignCenter
    TextAlignEnd
)

type TruncateMode uint8
const (
    TruncateClip TruncateMode = iota
    TruncateEllipsis
)

type Orientation uint8
const (
    OrientationHorizontal Orientation = iota
    OrientationVertical
)
```

### `Box`

```go
type BorderStyle uint8

const (
    BorderNone BorderStyle = iota
    BorderSingle
    BorderRounded
    BorderDouble
    BorderHeavy
)

type BoxProps struct {
    Width, Height        omnitui.Size
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
    Padding              omnitui.Spacing
    Gap                  int
    Direction            Direction
    Align                Align
    Justify              Justify
    Wrap                 bool
    Clip                 bool
    Border               BorderStyle
    Style                omnitui.Style

    Focusable bool
    Disabled  bool

    OnKey       omnitui.EventHandler[omnitui.KeyEvent]
    OnTextInput omnitui.EventHandler[omnitui.TextInputEvent]
    OnPaste     omnitui.EventHandler[omnitui.PasteEvent]
    OnFocus     omnitui.EventHandler[omnitui.FocusEvent]
    OnBlur      omnitui.EventHandler[omnitui.BlurEvent]
    OnPress     omnitui.EventHandler[omnitui.PressEvent]
    OnMouse     omnitui.EventHandler[omnitui.MouseEvent]
    OnWheel     omnitui.EventHandler[omnitui.WheelEvent]
    OnResize    omnitui.EventHandler[omnitui.ResizeEvent]
    OnMessage   omnitui.EventHandler[omnitui.MessageEvent]
}

func Box(props BoxProps, children ...omnitui.Element) omnitui.Element
```

`OnResize` e `OnMessage` só são chamados quando a `Box` é o nó host raiz. Quando diferente de `BorderNone`, `Border` ocupa uma célula em cada um dos quatro lados.

### `Row`

```go
type RowProps struct {
    Width, Height        omnitui.Size
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
    Padding              omnitui.Spacing
    Gap                  int
    Align                Align
    Justify              Justify
    Wrap                 bool
    Clip                 bool
    Style                omnitui.Style
}

func Row(props RowProps, children ...omnitui.Element) omnitui.Element
```

### `Column`

```go
type ColumnProps struct {
    Width, Height        omnitui.Size
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
    Padding              omnitui.Spacing
    Gap                  int
    Align                Align
    Justify              Justify
    Clip                 bool
    Style                omnitui.Style
}

func Column(props ColumnProps, children ...omnitui.Element) omnitui.Element
```

### `Text`

```go
type TextProps struct {
    Content  string
    Style    omnitui.Style
    Wrap     TextWrap
    Align    TextAlign
    MaxLines int
    Truncate TruncateMode
}

func Text(props TextProps) omnitui.Element
```

`MaxLines == 0` significa sem limite. Wrap e truncamento operam sobre graphemes e largura visual.

### `Button`

```go
type ButtonProps struct {
    Label         string
    Disabled      bool
    Style         omnitui.Style
    FocusStyle    omnitui.Style
    DisabledStyle omnitui.Style

    OnKey   omnitui.EventHandler[omnitui.KeyEvent]
    OnFocus omnitui.EventHandler[omnitui.FocusEvent]
    OnBlur  omnitui.EventHandler[omnitui.BlurEvent]
    OnPress omnitui.EventHandler[omnitui.PressEvent]
    OnMouse omnitui.EventHandler[omnitui.MouseEvent]
}

func Button(props ButtonProps) omnitui.Element
```

`Button` é focável por padrão. `Enter`, `Space` e clique esquerdo completo produzem `PressEvent` quando habilitado.

### `Input`

```go
type InputProps struct {
    Value       string
    Placeholder string
    Width       omnitui.Size
    Disabled    bool
    ReadOnly    bool
    Mask        rune
    MaxLength   int
    Style       omnitui.Style
    FocusStyle  omnitui.Style

    OnChange    omnitui.EventHandler[omnitui.ValueChangeEvent]
    OnSubmit    omnitui.EventHandler[omnitui.SubmitEvent]
    OnKey       omnitui.EventHandler[omnitui.KeyEvent]
    OnTextInput omnitui.EventHandler[omnitui.TextInputEvent]
    OnPaste     omnitui.EventHandler[omnitui.PasteEvent]
    OnFocus     omnitui.EventHandler[omnitui.FocusEvent]
    OnBlur      omnitui.EventHandler[omnitui.BlurEvent]
    OnMouse     omnitui.EventHandler[omnitui.MouseEvent]
}

func Input(props InputProps) omnitui.Element
```

`Input` é controlado por `Value`; `OnChange` apenas propõe um novo valor. `MaxLength` conta graphemes. `Mask` altera somente a pintura. Clique esquerdo posiciona o cursor no grapheme visual mais próximo.

### `Tabs`

```go
type TabItem struct {
    Key      string
    Label    string
    Content  omnitui.Element
    Disabled bool
}

type TabsProps struct {
    Items       []TabItem
    ActiveKey   string
    Orientation Orientation
    Style       omnitui.Style
    ActiveStyle omnitui.Style
    OnChange    omnitui.EventHandler[omnitui.ValueChangeEvent]
}

func Tabs(props TabsProps) omnitui.Element
```

Chaves devem ser únicas. `ActiveKey == ""` usa a primeira aba habilitada; uma chave inexistente ou desabilitada é erro de props. Clique esquerdo em um cabeçalho propõe sua chave por `OnChange`.

### `List`

```go
type ScrollbarMode uint8

const (
    ScrollbarAuto ScrollbarMode = iota
    ScrollbarAlways
    ScrollbarHidden
)

type ListProps struct {
    SelectedKey   string
    Height        omnitui.Size
    Gap           int
    Disabled      bool
    Wrap          bool
    ScrollPadding int
    Scrollbar     ScrollbarMode
    Empty         omnitui.Element
    Style         omnitui.Style
    SelectedStyle omnitui.Style

    OnChange   omnitui.EventHandler[omnitui.ValueChangeEvent]
    OnActivate omnitui.EventHandler[omnitui.ActivateEvent]
    OnMouse    omnitui.EventHandler[omnitui.MouseEvent]
    OnWheel    omnitui.EventHandler[omnitui.WheelEvent]
}

func List(props ListProps, items ...omnitui.Element) omnitui.Element
```

Cada item direto deve possuir `WithKey`. `List` é controlada por `SelectedKey`; clique esquerdo propõe o item e wheel move somente a viewport. Scroll e navegação detalhados ficam em [COMPONENTS.md](COMPONENTS.md#scroll).

## 9. Eventos — `omnitui`

### Contrato dos handlers

```go
type EventResult uint8

const (
    Propagate EventResult = iota
    Consume
)

type EventHandler[E any] func(event E) EventResult
```

- `Propagate` envia o evento ao próximo ancestral elegível.
- `Consume` interrompe bubbling e impede o comportamento padrão associado.
- Handler ausente equivale a `Propagate`.
- Handlers rodam na goroutine do runtime e devem retornar rapidamente.
- Eventos são valores imutáveis para o consumidor.

Para eventos sem propagação, o retorno é ignorado, mas a assinatura permanece uniforme.

### Eventos suportados no MVP

| Evento | Origem | Alvo inicial | Propagação |
|---|---|---|---|
| `KeyEvent` | Tecla ou sequência ANSI | Elemento focado | Alvo até ancestrais |
| `TextInputEvent` | Entrada imprimível normalizada | `Input` focado | Alvo até ancestrais |
| `PasteEvent` | Bracketed paste | `Input` focado | Alvo até ancestrais |
| `MouseEvent` | Movimento, botão, entrada ou saída do ponteiro | Host sob o ponteiro ou capturador | Alvo até ancestrais, exceto enter/leave |
| `WheelEvent` | Roda ou gesto de scroll do terminal | Host sob o ponteiro | Alvo até ancestrais |
| `FocusEvent` | Elemento recebe foco | Novo foco | Não propaga |
| `BlurEvent` | Elemento perde foco | Foco anterior | Não propaga |
| `PressEvent` | Ativação de controle | `Button` ou `Box` pressionável | Alvo até ancestrais |
| `ValueChangeEvent` | `Input`, `Tabs` ou `List` propõe valor | Builtin emissor | Não propaga |
| `SubmitEvent` | `Enter` em `Input` | `Input` focado | Não propaga |
| `ActivateEvent` | `Enter` em item de `List` | `List` focada | Não propaga |
| `ResizeEvent` | Dimensão do terminal muda | Host raiz | Não propaga |
| `MessageEvent` | `App.Dispatch` | Host raiz | Não propaga |

### Teclado

```go
type Key uint16

const (
    KeyRune Key = iota
    KeyEnter
    KeyEscape
    KeyTab
    KeyBacktab
    KeyBackspace
    KeyDelete
    KeyInsert
    KeyUp
    KeyDown
    KeyLeft
    KeyRight
    KeyHome
    KeyEnd
    KeyPageUp
    KeyPageDown
    KeyF1
    KeyF2
    KeyF3
    KeyF4
    KeyF5
    KeyF6
    KeyF7
    KeyF8
    KeyF9
    KeyF10
    KeyF11
    KeyF12
)

type Modifiers uint8

const (
    ModCtrl Modifiers = 1 << iota
    ModAlt
    ModShift
)

type KeyEvent struct {
    Key       Key
    Rune      rune
    Modifiers Modifiers
    Repeat    bool
}
```

Terminais não fornecem `KeyUp` de forma portátil. Modificadores e repetição só são reportados quando distinguíveis na entrada recebida.

Comportamentos padrão após um `KeyEvent` não consumido:

- `Tab` e `Backtab` movem o foco;
- `Enter` e `Space` geram `PressEvent` em controles pressionáveis;
- `Ctrl+C` encerra `Run` com `ErrInterrupted`.

### Foco

```go
type FocusCause uint8

const (
    ProgrammaticFocus FocusCause = iota
    ForwardTraversal
    BackwardTraversal
    ElementRemoved
)

type FocusEvent struct {
    Cause FocusCause
}

type BlurEvent struct {
    Cause FocusCause
}
```

`FocusEvent` e `BlurEvent` não propagam.

### Press

```go
type PressSource uint8

const (
    KeyboardEnter PressSource = iota
    KeyboardSpace
    MouseLeft
    ProgrammaticPress
)

type PressEvent struct {
    Source PressSource
}
```

Controles desabilitados não recebem nem propagam `PressEvent`.

### Mouse e wheel

```go
type MouseAction uint8

const (
    MouseMove MouseAction = iota
    MouseDown
    MouseUp
    MouseEnter
    MouseLeave
)

type MouseButton uint8

const (
    MouseButtonNone MouseButton = iota
    MouseButtonLeft
    MouseButtonMiddle
    MouseButtonRight
)

type MouseButtons uint8

const (
    MouseLeftPressed MouseButtons = 1 << iota
    MouseMiddlePressed
    MouseRightPressed
)

type MouseEvent struct {
    Action    MouseAction
    Button    MouseButton
    Buttons   MouseButtons
    X, Y      int
    LocalX    int
    LocalY    int
    Modifiers Modifiers
}

type WheelEvent struct {
    X, Y      int
    LocalX    int
    LocalY    int
    DeltaX    int
    DeltaY    int
    Modifiers Modifiers
}
```

- Coordenadas são zero-based; `X`/`Y` são relativas à tela. O runtime entrega uma cópia por handler com `LocalX`/`LocalY` relativos ao nó cujo handler está em execução, inclusive durante bubbling.
- `MouseMove`, `MouseDown`, `MouseUp` e `WheelEvent` propagam do alvo aos ancestrais.
- `MouseEnter` e `MouseLeave` são derivados comparando os caminhos de ancestrais. Cada nó que entrou ou saiu recebe seu próprio evento, sem bubbling; assim, uma `Box` detecta hover mesmo quando um filho é o alvo mais profundo.
- `Buttons` descreve os botões mantidos pressionados durante move; `Button` identifica o botão que mudou em down/up.
- `DeltaY < 0` rola para cima e `DeltaY > 0` para baixo; `DeltaX < 0` rola para a esquerda e `DeltaX > 0` para a direita.
- O backend normaliza wheel em linhas lógicas; a magnitude pode ser maior que 1 quando o terminal reportar vários passos.

Sem captura, hit testing percorre hosts do último pintado para o primeiro, respeita o recorte acumulado dos ancestrais e escolhe o nó visível mais profundo. O alvo não depende de a célula conter um grapheme: todo o retângulo de layout participa. Durante captura, move/up são entregues ao capturador; enter/leave continuam sendo calculados a partir do host realmente sob o ponteiro.

No `MouseDown`, o alvo recebe captura automática até o `MouseUp` correspondente. Eventos continuam chegando a ele mesmo fora de seu retângulo. A captura é cancelada se o alvo for desmontado ou desabilitado.

Comportamentos padrão não consumidos:

- clique esquerdo dá foco a um alvo focável;
- down esquerdo seguido de up ainda dentro do mesmo controle pressionável gera `PressEvent{Source: MouseLeft}`;
- clique em cabeçalho de `Tabs` propõe sua chave por `ValueChangeEvent`;
- clique em item de `List` propõe sua seleção;
- wheel sobre `List` desloca a viewport sem alterar `SelectedKey`.

O MVP habilita o protocolo SGR extended mouse e rastreamento de movimento. Double click, triple click, drag semântico, seleção por arraste e manipulação da scrollbar ficam fora do MVP; aplicações ainda podem interpretar a sequência bruta de down/move/up.

### Texto e paste

```go
type TextInputEvent struct {
    Text string
}

type PasteEvent struct {
    Text string
}
```

Para entrada imprimível, o runtime entrega `KeyEvent`, depois `TextInputEvent` e, se ambos permitirem o comportamento padrão, o `Input` propõe `ValueChangeEvent`. Paste permanece em um único evento e respeita o limite de entrada do backend.

### Mudança, submit e ativação

```go
type ChangeSource uint8

const (
    ChangeKeyboard ChangeSource = iota
    ChangePaste
    ChangeProgrammatic
)

type ValueChangeEvent struct {
    Previous string
    Value    string
    Source   ChangeSource
}

type SubmitEvent struct {
    Value string
}

type ActivateEvent struct {
    Key    string
    Source PressSource
}
```

`ValueChangeEvent` é uma proposta: não altera props automaticamente. `SubmitEvent` não limpa o input e `ActivateEvent` não muda a seleção.

### Resize e mensagens

```go
type ResizeEvent struct {
    Width  int
    Height int
}

type MessageEvent struct {
    Value any
}
```

Resizes pendentes podem ser coalescidos para o tamanho mais recente. Mensagens preservam ordem e não são coalescidas.

### Ordem de processamento

1. Normalizar a entrada do backend.
2. Resolver o alvo por foco, raiz, captura de mouse ou hit testing.
3. Executar o handler do alvo.
4. Propagar enquanto o resultado for `Propagate`.
5. Aplicar o comportamento padrão se o evento não foi consumido.
6. Drenar atualizações de estado produzidas pelos handlers.
7. Reconciliar e gerar no máximo um frame.

### Eventos fora do MVP

- `TickEvent`;
- `KeyUp`;
- `DoubleClickEvent` e `DragEvent` semânticos;
- lifecycle `Mount`, `Update` e `Unmount`.

Lifecycle pertencerá à futura API de efeitos e cleanup, não ao sistema de eventos de input.

## 10. Matriz de handlers por componente

| Componente | Handlers públicos |
|---|---|
| `Box` | `OnKey`, `OnTextInput`, `OnPaste`, `OnFocus`, `OnBlur`, `OnPress`, `OnMouse`, `OnWheel`, `OnResize`, `OnMessage` |
| `Button` | `OnKey`, `OnFocus`, `OnBlur`, `OnPress`, `OnMouse` |
| `Input` | `OnKey`, `OnTextInput`, `OnPaste`, `OnFocus`, `OnBlur`, `OnMouse`, `OnChange`, `OnSubmit` |
| `Tabs` | `OnChange` |
| `List` | `OnMouse`, `OnWheel`, `OnChange`, `OnActivate` |
| `Row`, `Column`, `Text` | Nenhum handler direto |

Para tornar `Row` ou `Column` interativo, use uma `Box` configurada como focável ou crie um componente composto que renderize uma superfície interativa.

## 11. Erros de uso

As situações abaixo são erros de programação e devem incluir o caminho do componente:

- chave duplicada entre irmãos;
- tipo de estado incompatível;
- atualização de estado durante `Render`;
- children em componente folha;
- props com tamanhos negativos;
- aba ativa inexistente ou desabilitada;
- item de `List` sem chave;
- atributo simultaneamente presente em `Attributes` e `ClearAttributes`.

Erros de I/O, cancelamento e interrupção saem de `App.Run`.
