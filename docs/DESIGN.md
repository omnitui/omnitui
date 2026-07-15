# OmniTUI — plano de design

A referência pública de tipos, funções, estilos e eventos está em [API.md](API.md).

## 1. Objetivo

Construir um framework TUI em Go, inspirado no modelo mental do React, no qual:

- a interface é uma árvore declarativa de elementos;
- componentes recebem props, contexto e filhos;
- cada instância montada de componente possui estado próprio;
- todo componente renderiza exatamente um elemento (um `Fragment` cobre vários filhos);
- alterações de estado disparam uma nova renderização;
- a reconciliação preserva ou descarta estado de forma previsível;
- o pacote público `components` inclui `Row`, `Column`, `Text`, `Input`, `Tabs` e `List` como builtins oficiais;
- a tela é atualizada por diferença entre buffers, sem redesenhar o terminal inteiro.

O primeiro marco deve provar esse modelo com uma API pequena. Hooks, efeitos, animações, renderização concorrente e uma imitação completa de CSS ficam fora do MVP.

## 2. Premissas e limites

1. A primeira plataforma suportada será terminal Unix (Linux e macOS) com ANSI/VT100.
2. “Do zero” significa não encapsular Bubble Tea, tview ou outro framework TUI. Utilitários pequenos e focados, como `golang.org/x/term` para modo raw e uma biblioteca de largura Unicode, são aceitáveis.
3. Todos os renders e mutações da árvore acontecerão em uma única goroutine, controlada pelo runtime. Outras goroutines poderão apenas enfileirar mensagens.
4. Props e elementos serão tratados como valores imutáveis. O runtime será o dono do estado montado.
5. A primeira implementação reconciliará toda a árvore lógica após uma atualização. Otimizações por subárvore só serão adicionadas se medições justificarem.
6. O layout inicial será um subconjunto pequeno de flexbox: linha/coluna, tamanho, padding, gap, alinhamento e clipping.
7. Windows, IME, efeitos assíncronos e acessibilidade avançada serão extensões posteriores; mouse SGR faz parte do MVP Unix.

## 3. Modelo mental

Há três representações diferentes e elas não devem ser misturadas:

```text
Elementos imutáveis          Instâncias montadas          Saída física
(o que o usuário quer)  ->   (identidade + estado)   ->   (buffer de células)

Component / Box / Text       componentInstance            Cell[x,y]
props / key / children       state / context / host       rune / width / style
```

- **Elemento:** descrição barata e descartável criada durante `Render`.
- **Instância:** objeto interno persistente enquanto tipo, posição e chave continuarem compatíveis. É aqui que o estado vive.
- **Nó host:** resultado sem componentes de usuário, contendo somente primitivas que o layout entende.
- **Buffer de células:** representação final da tela usada para produzir sequências ANSI mínimas.

Essa separação evita guardar estado dentro de `Element`, que é recriado em cada render, e evita acoplar componentes diretamente ao terminal.

## 4. API pública

A referência canônica dos pacotes `omnitui` e `omnitui/components` está em [API.md](API.md). Ela contém elementos, componentes, estado, contexto, runtime, geometria, estilos, builtins e eventos suportados.

Este documento conserva apenas decisões arquiteturais e exemplos que ajudam a explicar o modelo.

## 5. Exemplo de uso desejado

O exemplo assume `omnitui` como nome do pacote raiz e `components` como nome do pacote `omnitui/components`.

```go
type CounterProps struct {
    Label string
}

type CounterState struct {
    Value int
}

type Counter struct{}

func (Counter) InitialState(CounterProps) CounterState {
    return CounterState{}
}

func (Counter) Render(
    ctx omnitui.Context,
    props CounterProps,
    state CounterState,
    children omnitui.Children,
) omnitui.Element {
    return components.Column(
        components.ColumnProps{Gap: 1},
        components.Text(components.TextProps{
            Content: fmt.Sprintf("%s: %d", props.Label, state.Value),
        }),
        components.Button(components.ButtonProps{
            Label: "Incrementar",
            OnPress: func(event omnitui.PressEvent) omnitui.EventResult {
                omnitui.UpdateState(ctx, func(current CounterState) CounterState {
                    current.Value++
                    return current
                })
                return omnitui.Consume
            },
        }),
        omnitui.Fragment(children...),
    )
}

var CounterType = omnitui.Define[CounterProps, CounterState]("Counter", Counter{})

root := omnitui.Create(
    CounterType,
    CounterProps{Label: "Cliques"},
    components.Text(components.TextProps{
        Content: "Filho recebido pelo componente",
    }),
)
```

Esse exemplo define o contrato mínimo que os primeiros testes de integração devem tornar possível.

## 6. Reconciliação e identidade

Ao receber uma nova árvore de elementos, o reconciliador compara cada elemento novo com a instância montada anterior:

1. Se tipo e chave forem compatíveis, reutiliza a instância e seu estado.
2. Se tipo ou chave mudarem, desmonta a instância anterior e monta uma nova com `InitialState`.
3. Sem chave, identidade é determinada pela posição entre irmãos.
4. Com chave, a busca acontece no conjunto de irmãos e permite reordenação sem perder estado.
5. Chaves duplicadas entre irmãos geram erro com o caminho da árvore.
6. Props e filhos sempre são substituídos pelos valores do render mais recente.

Para uma instância de componente reutilizada:

1. aplicar atualizações de estado pendentes;
2. construir o contexto herdado;
3. chamar `Render` com props, estado e filhos atuais;
4. reconciliar o elemento retornado com a subárvore anterior.

Para um nó host reutilizado:

1. atualizar props;
2. reconciliar filhos;
3. produzir ou atualizar o nó usado pelo layout.

O MVP não terá lifecycle público. Internamente, a desmontagem invalida o dispatcher da instância e remove seus handlers/foco. Uma API de cleanup só deve nascer junto com efeitos assíncronos, em uma fase posterior.

### Pseudocódigo

```text
reconcile(parent, oldInstance, newElement, inheritedContext):
    if newElement is None:
        unmount(oldInstance)
        return nil

    if identity(oldInstance) != identity(newElement):
        unmount(oldInstance)
        oldInstance = mount(newElement)

    update oldInstance props and children

    if oldInstance is Provider:
        nextContext = inheritedContext.with(key, value)
        reconcile its child with nextContext

    if oldInstance is Component:
        apply queued state updates
        output = component.Render(contextFor(oldInstance), props, state, children)
        reconcile oldInstance.rendered with output

    if oldInstance is Host:
        reconcileChildren(oldInstance, newElement.children)

    return oldInstance
```

## 7. Pipeline de renderização

```text
input / resize / SetState
          |
          v
      fila do runtime -- agrupa atualizações
          |
          v
  render + reconciliação -- preserva identidade e estado
          |
          v
      árvore host -- somente Box, Text e hosts internos de interação
          |
          v
        layout -- posições, tamanhos e clipping
          |
          v
         paint -- buffer traseiro de Cell
          |
          v
     diff de buffers -- sequências ANSI para células alteradas
          |
          v
        terminal
```

### Buffer de tela

```go
type Cell struct {
    Grapheme string
    Width    int
    Style    Style
}
```

O renderer mantém buffer frontal e traseiro. Depois do paint, ele agrupa células alteradas por linha, reduz movimentos de cursor e só emite mudanças de estilo quando necessário. Ao final, troca os buffers.

Unicode exige tratar largura visual, caracteres combinantes e células de continuação. Isso deve ser coberto desde a primeira versão do buffer, mesmo que o parser de entrada comece apenas com teclado básico.

## 8. Layout, estilo e componentes builtin

### 8.1 Catálogo de componentes

O pacote `omnitui/components` entregará `Row`, `Column`, `Text`, `Input`, `Tabs` e `List` como builtins oficiais. Seus contratos, props, comportamentos e exemplos de utilização ficam em [COMPONENTS.md](COMPONENTS.md).

`Box` e `Button` permanecem no mesmo pacote como blocos visuais de nível mais baixo; `Fragment` e `None` pertencem ao núcleo `omnitui`. `Box` e `Text` criam hosts por meio da fronteira opaca `internal/core`; os demais builtins usam a mesma API de `Component` disponível aos usuários. `Input` usa um host interno não exportado para cursor e edição.

### 8.2 Motor de layout

O algoritmo terá duas passagens:

1. **Measure, de baixo para cima:** cada nó calcula seu tamanho desejado dentro das restrições recebidas.
2. **Layout, de cima para baixo:** o pai distribui espaço e define o retângulo final de cada filho.

Tipos centrais internos:

```go
type Constraints struct {
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
}

type Rect struct {
    X, Y, Width, Height int
}
```

Não será implementado CSS genérico. Cada prop suportada terá semântica documentada e teste próprio.

## 9. Eventos, foco e concorrência

Tipos públicos, handlers, teclas, eventos suportados e regras de propagação estão em [API.md](API.md).

O loop principal será aproximadamente:

```text
for app is running:
    wait for input, resize, external message, state update, or cancellation
    drain queued work
    route events
    reconcile if invalidated
    layout and paint if invalidated or resized
    flush screen diff
```

Somente a goroutine do runtime toca em instâncias, layout e buffers. Leitores de terminal e produtores externos publicam valores em canais. Handlers são executados em série; atualizações síncronas produzidas pelo mesmo evento são drenadas antes da reconciliação e aparecem juntas no próximo frame.

O runtime mantém a ordem de foco derivada da árvore host visível. Desmontar ou desabilitar o nó focado escolhe o próximo candidato válido; resize invalida layout e paint sem reinicializar componentes.

Eventos de mouse usam a árvore host já posicionada para hit testing. A busca percorre a ordem de pintura do topo para o fundo, respeita retângulos de clipping e escolhe o nó visível mais profundo. Mouse down estabelece captura temporária até mouse up; desmontar o alvo cancela a captura. O caminho anterior sob o ponteiro é comparado ao novo para derivar enter e leave sem armazenar estado em componentes de usuário.

## 10. Organização inicial do código

Uma prévia completa da árvore de diretórios, responsabilidade dos arquivos, direção de dependências e criação por fase está em [STRUCTURE.md](STRUCTURE.md).

O código expõe dois pacotes: `omnitui` para o runtime e `omnitui/components` para os builtins. `components` depende do núcleo; o núcleo nunca importa o catálogo. Ambos compartilham apenas a representação opaca de elementos em `internal/core`.

Extrações para `internal/` só devem ocorrer quando uma fronteira estiver estável e houver benefício concreto de isolamento. O backend de terminal é a primeira fronteira provável, pois permitirá um backend headless para testes e, futuramente, Windows.

Interfaces internas que valem desde cedo:

```go
type Backend interface {
    Size() (width, height int, err error)
    Events() <-chan Event
    Write([]byte) error
    Close() error
}

type Clock interface {
    Now() time.Time
}
```

`Backend` viabiliza testes determinísticos. `Clock` só deve ser introduzido quando animações ou timers realmente entrarem no produto.

## 11. Dependências

Orçamento inicial de dependências:

- `golang.org/x/term`: modo raw e tamanho do terminal;
- uma biblioteca pequena de graphemes/largura Unicode, escolhida após um spike comparativo;
- nenhuma dependência de framework TUI, layout ou gerenciamento de estado.

Se “do zero” precisar significar literalmente apenas biblioteca padrão, será necessário manter código específico por sistema operacional e tabelas Unicode próprias. Isso aumenta bastante o custo sem melhorar o modelo de componentes; por isso não é a recomendação inicial.

## 12. Tratamento de erros e diagnóstico

- Erros de I/O e cancelamento saem de `Run`.
- Erros de uso — chave duplicada, tipo de estado incorreto ou atualização durante render — incluem o caminho de componentes.
- Um panic de componente deve primeiro restaurar o terminal; depois pode ser propagado ou embrulhado conforme a política escolhida para `Run`.
- Uma opção de diagnóstico futura poderá registrar causa de cada render, duração das fases e regiões alteradas da tela.
- Não haverá error boundary no MVP; primeiro deve existir um contrato claro de recuperação do terminal.

## 13. Estratégia de testes

O núcleo deve ser testável sem um terminal real.

### Unidade

- `Element` preserva props, chave e filhos sem mutação.
- `InitialState` é chamado uma vez por montagem.
- estado sobrevive a render com mesmo tipo e chave.
- estado é reinicializado quando tipo ou chave muda.
- reordenação com chaves move instâncias e preserva seus estados.
- props e filhos novos chegam ao render reutilizado.
- o provider mais próximo vence e seu valor não vaza para irmãos.
- atualizações funcionais são aplicadas em ordem e agrupadas por frame.
- o parser reconhece todas as teclas declaradas no MVP e preserva modificadores disponíveis.
- o parser reconhece SGR mouse, botões, movimento, release, wheel, coordenadas e modificadores.
- `Consume` impede bubbling e comportamento padrão; `Propagate` mantém ambos.
- foco e blur são emitidos uma vez, na ordem correta, sem propagação.
- `Enter` e `Space` geram `PressEvent` somente quando o `KeyEvent` não é consumido.
- hit testing respeita clipping e ordem de pintura; captura entrega move/up ao alvo original.
- enter/leave são derivados uma vez por transição e clique esquerdo gera `PressEvent` somente com down/up compatíveis.
- texto, paste, mudança, submit e ativação seguem a ordem declarada em [API.md](API.md).
- resizes são coalescidos e mensagens externas preservam ordem.
- resize preserva estado e recalcula layout.
- measure/layout respeita constraints, clipping e Unicode.
- diff gera ANSI apenas para células alteradas.
- `Row` e `Column` convertem suas props para `Box` sem alterar identidade dos filhos.
- `Text` mede, quebra e trunca por largura visual.
- `Input` preserva cursor local, mantém valor controlado e trata edição, paste e submit.
- `Tabs` valida chaves, ignora abas desabilitadas e aceita seleção por teclado ou clique.
- `List` exige chaves, preserva seleção por chave e mantém o item selecionado visível.
- scroll da `List` limita offsets, preserva âncora em reordenação e funciona com itens de alturas diferentes.
- wheel desloca a `List` sem alterar `SelectedKey` e propaga quando não há mais espaço para rolar.

### Integração headless

Um `TestBackend` recebe eventos sintéticos e captura frames:

1. montar o contador do exemplo;
2. conferir o primeiro frame;
3. enviar `Tab` e `Enter`;
4. conferir texto incrementado e foco;
5. reordenar componentes com chaves;
6. confirmar que seus estados acompanharam as chaves;
7. editar e submeter um `Input` controlado;
8. navegar entre `Tabs` e itens de `List` por teclado;
9. clicar em `Button`, `Input`, `Tabs` e `List` com coordenadas sintéticas;
10. rolar uma `List` por wheel e conferir offset, clipping e seleção;
11. cancelar e verificar fechamento do backend.

### Terminal real

Testes em pseudo-terminal validam modo raw, ativação/desativação do protocolo SGR mouse, resize, restauração após erro e sequências ANSI. Snapshots são úteis para árvore host e buffers; o comportamento do reconciliador deve preferir asserções semânticas, menos frágeis.

## 14. Plano incremental

### Fase 0 — contrato executável

Entregas:

- inicializar módulo Go e CI básica;
- escrever o exemplo `Counter` como teste de compilação;
- definir `Element`, `Component`, props, estado, contexto e filhos;
- criar backend headless mínimo.

Critério de conclusão: o exemplo compila e monta uma árvore inspecionável, ainda sem terminal.

### Fase 1 — reconciliação e estado

Entregas:

- montar, atualizar e desmontar instâncias;
- implementar identidade posicional e por chave;
- implementar fila de `SetState`/`UpdateState`;
- implementar provider/consumer de contexto;
- adicionar diagnóstico de caminhos e chaves duplicadas.

Critério de conclusão: todos os testes de identidade, estado, props, filhos e contexto passam em memória.

### Fase 2 — árvore host e layout

Entregas:

- `Text`, `Box`, `Row`, `Column`, `Fragment` e `None`;
- measure/layout com row, column, constraints, padding, gap e clipping;
- buffer de células com Unicode e estilos;
- snapshots determinísticos da tela.

Critério de conclusão: árvores headless produzem buffers corretos em diferentes tamanhos.

### Fase 3 — terminal interativo

Entregas:

- backend Unix, modo raw e alternate screen;
- parser de teclas, SGR mouse, wheel e resize;
- ativação de mouse tracking e restauração dos modos do terminal;
- diff ANSI frontal/traseiro;
- restauração robusta do terminal.

Critério de conclusão: o contador roda em terminal real, redimensiona sem perder estado e sempre restaura o terminal ao sair.

### Fase 4 — interação

Entregas:

- foco, bubbling e consumo de eventos;
- `KeyEvent`, `MouseEvent`, `WheelEvent`, `FocusEvent`, `BlurEvent`, `PressEvent`, `TextInputEvent`, `PasteEvent`, `ResizeEvent` e `MessageEvent`;
- hit testing, hover path e captura de mouse;
- comportamentos padrão canceláveis e despacho ordenado;
- `Button` acessível por teclado e mouse;
- testes end-to-end com pseudo-terminal;
- medição de tempo de render/layout/paint.

Critério de conclusão: interação por teclado e mouse é determinística, eventos atingem o alvo correto sob clipping e um frame comum não reescreve células inalteradas.

### Fase 5 — builtins com estado e seleção

Entregas:

- `Input` controlado, cursor, edição, paste, submit e posicionamento por clique;
- `Tabs` controlado, navegação de cabeçalhos, clique e painel ativo;
- `List` controlada, viewport, navegação, clique, ativação, scrollbar, wheel e scroll automático;
- `ValueChangeEvent`, `SubmitEvent` e `ActivateEvent`;
- documentação e exemplos de composição para todos os builtins.

Critério de conclusão: `Row`, `Column`, `Text`, `Input`, `Tabs` e `List` são exportados por `omnitui/components`, não criam dependência inversa no runtime e passam pelos mesmos testes de reconciliação que componentes de usuário.

### Fase 6 — endurecimento antes de expandir

Entregas:

- race detector, fuzzing do parser ANSI e testes de panic;
- documentação pública e dois exemplos maiores;
- benchmarks com árvores profundas, listas com chaves e tela cheia;
- decisão baseada em medidas sobre memoização ou render por subárvore.

Critério de conclusão: API mínima estabilizada e gargalos conhecidos por benchmark, não por suposição.

## 15. Decisões deliberadamente adiadas

- hooks (`UseState`, `UseEffect`);
- lifecycle e efeitos com cleanup;
- componentes assíncronos ou suspense;
- render concorrente;
- memoização pública;
- portals e overlays fora da árvore normal;
- double click, drag semântico, arraste de scrollbar, acesso direto ao clipboard, IME e Windows;
- virtualização de listas;
- animações e timers;
- markup ou DSL própria.

Cada item só deve entrar com um caso de uso, semântica definida e teste. O núcleo não deve antecipar todas as capacidades do React: a inspiração principal é árvore declarativa, fluxo de dados, identidade e reconciliação previsível.

## 16. Critérios de sucesso do MVP

O MVP estará pronto quando for possível demonstrar, com testes e um exemplo executável, que:

1. um componente recebe props, contexto e filhos tipados;
2. duas instâncias do mesmo componente mantêm estados independentes;
3. uma atualização de estado causa novo render sem bloquear ou corromper o loop;
4. tipo, posição e chave determinam corretamente se o estado é preservado;
5. componentes compõem primitivas e outros componentes recursivamente;
6. resize e eventos não reinicializam a árvore;
7. somente diferenças do buffer são escritas no terminal;
8. o terminal é restaurado em toda forma de saída;
9. `go test -race ./...` passa;
10. `Row`, `Column`, `Text`, `Input`, `Tabs` e `List` são exportados exclusivamente por `omnitui/components`;
11. mouse SGR, hit testing, captura, press por clique e wheel funcionam no backend real e headless;
12. a API do exemplo `Counter` permanece pequena e compreensível.

Esses critérios formam a linha de corte. Recursos que não ajudam diretamente a cumpri-los devem esperar a primeira versão funcionar de ponta a ponta.
