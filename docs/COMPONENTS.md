# OmniTUI — componentes builtin

Este documento demonstra os componentes oficiais `Row`, `Column`, `Text`, `Input`, `Tabs` e `List`, exportados pelo pacote público `omnitui/components`. A referência de assinaturas e props está em [API.md](API.md); o modelo de renderização está em [DESIGN.md](DESIGN.md).

## 1. Convenções

- Todos os builtins recebem props.
- `Row`, `Column` e `List` recebem filhos.
- `Tabs` recebe seus painéis como `Element` dentro de `TabItem`.
- `Text` e `Input` são folhas e não recebem filhos.
- `Input`, `Tabs` e `List` são controlados: o valor público vem das props e eventos propõem alterações ao componente pai.
- Estado interno guarda somente detalhes de interação, como cursor, foco local e offset de scroll.
- Handlers retornam `omnitui.Propagate` ou `omnitui.Consume`, conforme o contrato de eventos.
- `omnitui.Cells(n)` cria um `Size` fixo em células; `omnitui.All(n)` cria espaçamento igual nos quatro lados.
- Exemplos usam `omnitui` para o núcleo e `components` para os builtins; representam trechos de um método `Render` e omitem tratamento de erros.

```go
import (
    omnitui "github.com/viniciusfonseca/omnitui"
    components "github.com/viniciusfonseca/omnitui/components"
)
```

| Componente | Finalidade | Recebe filhos | Estado interno |
|---|---|---:|---|
| `Row` | Layout horizontal | Sim | Não |
| `Column` | Layout vertical | Sim | Não |
| `Text` | Texto estático | Não | Não |
| `Input` | Edição de texto em uma linha | Não | Cursor e scroll horizontal |
| `Tabs` | Navegação entre painéis | Via `TabItem.Content` | Cabeçalho focado |
| `List` | Lista selecionável e rolável | Sim | Foco e offset vertical |

## 2. `Row`

Organiza filhos horizontalmente. É implementado sobre `Box` com direção horizontal.

Assinatura e props: [API.md — `Row`](API.md#row).

Por padrão, mede a soma das larguras dos filhos, usa a maior altura e não quebra linha. `Justify` distribui espaço no eixo horizontal; `Align` posiciona os filhos no eixo vertical.

### Exemplo

```go
return components.Row(
    components.RowProps{
        Gap:     1,
        Align:   components.AlignCenter,
        Justify: components.JustifyEnd,
    },
    components.Text(components.TextProps{Content: "Salvar alterações?"}),
    components.Button(components.ButtonProps{
        Label: "Cancelar",
        OnPress: func(event omnitui.PressEvent) omnitui.EventResult {
            cancel()
            return omnitui.Consume
        },
    }),
    components.Button(components.ButtonProps{
        Label: "Salvar",
        OnPress: func(event omnitui.PressEvent) omnitui.EventResult {
            save()
            return omnitui.Consume
        },
    }),
)
```

## 3. `Column`

Organiza filhos verticalmente. Compartilha a semântica de `Row`, invertendo os eixos.

Assinatura e props: [API.md — `Column`](API.md#column).

`Row` e `Column` têm props separadas para manter a API legível, embora convertam internamente para a mesma estrutura de layout.

### Exemplo

```go
return components.Column(
    components.ColumnProps{
        Gap:     1,
        Padding: omnitui.All(1),
    },
    components.Text(components.TextProps{Content: "Perfil"}),
    components.Text(components.TextProps{Content: "Nome: Ada Lovelace"}),
    components.Text(components.TextProps{Content: "Cargo: Engenheira"}),
)
```

## 4. `Text`

Renderiza texto estático, não recebe foco e não aceita filhos.

Assinatura e props: [API.md — `Text`](API.md#text).

O conteúdo é segmentado em graphemes antes da medição. `MaxLines == 0` significa sem limite. Wrapping e truncamento respeitam largura visual, não quantidade de bytes ou runes.

### Exemplo

```go
return components.Text(components.TextProps{
    Content:  "Uma descrição longa que pode ocupar mais de uma linha.",
    Wrap:     components.WrapWord,
    MaxLines: 2,
    Truncate: components.TruncateEllipsis,
    Style: omnitui.Style{
        Foreground: omnitui.ANSI(omnitui.Cyan),
        Attributes: omnitui.Bold,
    },
})
```

## 5. `Input`

Campo de texto de uma linha, focável e controlado. O valor exibido sempre vem de `Value`; `OnChange` propõe um novo valor e o pai deve atualizar seu estado para aceitá-lo.

Assinatura e props: [API.md — `Input`](API.md#input).

O estado interno guarda cursor e deslocamento horizontal, nunca uma segunda cópia de `Value`. Inserção, Backspace, Delete, setas, Home, End, paste e submit com Enter fazem parte da primeira versão. Clique esquerdo posiciona o cursor no grapheme visual mais próximo; arrastar para selecionar texto fica fora do MVP. `MaxLength` conta graphemes. `Mask` altera somente a pintura.

### Exemplo

```go
type FormState struct {
    Name      string
    Submitted string
}

func renderNameInput(ctx omnitui.Context, state FormState) omnitui.Element {
    return components.Input(components.InputProps{
        Value:       state.Name,
        Placeholder: "Digite seu nome",
        MaxLength:   80,
        OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
            omnitui.UpdateState(ctx, func(current FormState) FormState {
                current.Name = event.Value
                return current
            })
            return omnitui.Consume
        },
        OnSubmit: func(event omnitui.SubmitEvent) omnitui.EventResult {
            omnitui.UpdateState(ctx, func(current FormState) FormState {
                current.Submitted = event.Value
                return current
            })
            return omnitui.Consume
        },
    })
}
```

## 6. `Tabs`

Exibe uma barra de abas e o painel ativo. A seleção é controlada por `ActiveKey`.

Assinatura e props: [API.md — `Tabs`](API.md#tabs).

Chaves devem ser únicas e estáveis. `ActiveKey == ""` exibe a primeira aba habilitada. Uma chave inexistente ou desabilitada é erro de props. Setas movem o foco; `Enter`, `Space` ou clique esquerdo em um cabeçalho propõem uma nova chave.

### Exemplo

```go
type ScreenState struct {
    ActiveTab string
}

func renderTabs(ctx omnitui.Context, state ScreenState) omnitui.Element {
    return components.Tabs(components.TabsProps{
        ActiveKey: state.ActiveTab,
        Items: []components.TabItem{
            {
                Key:   "overview",
                Label: "Visão geral",
                Content: components.Text(components.TextProps{
                    Content: "Resumo do projeto",
                }),
            },
            {
                Key:   "logs",
                Label: "Logs",
                Content: components.Text(components.TextProps{
                    Content: "Nenhum erro encontrado",
                }),
            },
        },
        OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
            omnitui.UpdateState(ctx, func(current ScreenState) ScreenState {
                current.ActiveTab = event.Value
                return current
            })
            return omnitui.Consume
        },
    })
}
```

## 7. `List`

Apresenta filhos como itens selecionáveis em uma viewport vertical. Cada filho direto deve ter `WithKey`; a chave identifica a seleção e preserva identidade durante reordenação.

Assinatura e props: [API.md — `List`](API.md#list).

Setas, Home, End, PageUp e PageDown propõem uma nova seleção. `Enter` emite `ActivateEvent`. O estado interno mantém foco e offset de scroll; `SelectedKey` permanece controlado pelo pai. `Empty` é renderizado quando não há itens.

### Scroll

O `List` cria uma viewport rolável quando `Height` resolve para uma altura finita e a soma da altura visual dos itens e gaps ultrapassa esse espaço. Com altura automática, a lista cresce para comportar os itens e não há scroll.

O offset é medido em **linhas do terminal**, não em índices. Isso permite itens com múltiplas linhas e alturas diferentes. Internamente, o estado usa a chave do primeiro item visível e o deslocamento dentro dele como âncora; assim, inserir ou reordenar itens antes da viewport não causa um salto desnecessário.

Valores de `ScrollbarMode`: [API.md — `List`](API.md#list).

- `ScrollbarAuto` ocupa uma coluna somente quando existe overflow.
- `ScrollbarAlways` reserva uma coluna mesmo quando todo o conteúdo cabe.
- `ScrollbarHidden` mantém o scroll, mas não desenha o indicador.
- `ScrollPadding` tenta manter essa quantidade de linhas livres acima e abaixo do item selecionado. Nas extremidades ou quando o item é maior que a viewport, a restrição é relaxada.

#### Navegação e scroll automático

- `Up` e `Down` propõem o item anterior ou seguinte. Quando o pai aceita a proposta em `SelectedKey`, a lista aplica o menor deslocamento necessário para torná-lo visível.
- `PageUp` e `PageDown` propõem o primeiro item elegível aproximadamente uma altura de viewport acima ou abaixo, descontando uma linha de sobreposição.
- `Home` propõe o primeiro item; quando aceito, move a viewport para o topo.
- `End` propõe o último item; quando aceito, move a viewport para o final.
- Com `Wrap`, ultrapassar uma extremidade propõe a extremidade oposta e, se aceita, desloca a viewport até ela.
- Alterar `SelectedKey` externamente revela o item correspondente no próximo layout. Essa é a forma declarativa de realizar scroll programático no MVP.
- Clique esquerdo em um item propõe sua chave como seleção e transfere o foco para a `List`.
- Wheel altera diretamente a âncora de scroll sem mudar `SelectedKey`; portanto, a seleção pode ficar temporariamente fora da viewport.
- `OnWheel` roda antes do comportamento padrão. `Consume` impede o scroll; no limite superior ou inferior, um wheel não consumido propaga para permitir containers roláveis aninhados no futuro.

Ao mudar a seleção, o item é considerado visível quando todo o seu retângulo cabe na viewport. Se ele for mais alto que a viewport, sua primeira linha é alinhada ao topo e o restante é recortado. Scroll manual por wheel não força esse invariant até a próxima mudança de `SelectedKey`.

#### Mudanças da árvore e resize

- Se itens forem inseridos ou reordenados, a âncora por chave preserva a região visível sempre que possível.
- Se o item âncora desaparecer, o runtime usa o item seguinte, depois o anterior, e por fim o topo.
- Se o item selecionado desaparecer, o componente emite uma proposta para a chave elegível mais próxima; ele não altera `SelectedKey` silenciosamente.
- Resize recalcula a viewport e limita o offset ao novo intervalo. Se a seleção estava visível antes do resize, ela continua visível; uma seleção já afastada por wheel não causa snap automático.
- Uma lista vazia zera o offset e renderiza `Empty` sem scrollbar.

Todos os filhos são reconciliados e medidos no MVP, inclusive os que estão fora da viewport; apenas o paint é recortado. Virtualização fica adiada até existirem benchmarks e uma API específica para itens sob demanda. Wheel integra o MVP; arraste da scrollbar e inércia permanecem adiados.

### Exemplo

```go
type ProjectState struct {
    SelectedProject string
    OpenProject     string
}

func renderProjects(ctx omnitui.Context, state ProjectState) omnitui.Element {
    return components.List(
        components.ListProps{
            SelectedKey:   state.SelectedProject,
            Height:        omnitui.Cells(4),
            Wrap:          true,
            ScrollPadding: 1,
            Scrollbar:     components.ScrollbarAuto,
            Empty: components.Text(components.TextProps{
                Content: "Nenhum projeto",
            }),
            OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
                omnitui.UpdateState(ctx, func(current ProjectState) ProjectState {
                    current.SelectedProject = event.Value
                    return current
                })
                return omnitui.Consume
            },
            OnActivate: func(event omnitui.ActivateEvent) omnitui.EventResult {
                omnitui.UpdateState(ctx, func(current ProjectState) ProjectState {
                    current.OpenProject = event.Key
                    return current
                })
                return omnitui.Consume
            },
        },
        components.Text(components.TextProps{Content: "OmniTUI"}).WithKey("omnitui"),
        components.Text(components.TextProps{Content: "CLI Tools"}).WithKey("cli-tools"),
        components.Text(components.TextProps{Content: "Experimentos"}).WithKey("labs"),
        components.Text(components.TextProps{Content: "Documentação"}).WithKey("docs"),
        components.Text(components.TextProps{Content: "Benchmarks"}).WithKey("benchmarks"),
        components.Text(components.TextProps{Content: "Arquivo"}).WithKey("archive"),
    )
}
```

## 8. Blocos de nível mais baixo

- `Box`: exportado por `components`; é um container configurável com direção, tamanho, padding, gap, alinhamento, borda, estilo e clipping.
- `Button`: exportado por `components`; é um controle focável com label e `OnPress`.
- `Fragment`: pertence a `omnitui` e agrupa elementos sem criar layout.
- `None`: pertence a `omnitui` e representa ausência de elemento.

`Row` e `Column` devem ser a escolha usual para layout. `Box` fica disponível quando a direção precisa ser dinâmica ou quando capacidades de nível mais baixo são necessárias.

## 9. Critérios de aceitação dos builtins

1. Todos são exportados por `omnitui/components` e usam o reconciliador do núcleo.
2. Props e filhos nunca são alterados internamente.
3. `Row` e `Column` preservam identidade e chaves de seus filhos.
4. `Text` mede, quebra e trunca corretamente graphemes de largura variável.
5. `Input` mantém cursor válido quando `Value` muda externamente.
6. `Tabs` valida chaves e nunca ativa uma aba desabilitada.
7. `List` preserva âncora e seleção por chave durante inserção e reordenação.
8. `List` revela o item após mudança de seleção e preserva sua visibilidade no resize quando ele já estava visível.
9. `List` limita o offset corretamente com itens variáveis, viewport pequena e lista vazia.
10. `Input`, `Tabs` e `List` respondem a clique; `List` responde a wheel sem alterar a seleção.
11. Todos os eventos seguem a ordem e propagação definidas em [API.md](API.md).
12. Todos funcionam no backend headless e possuem exemplos compilados como testes.
