# Discovery: custo de tokens do Claude Code

- Motivação:
  - US$ 98,39 gastos em 5 dias (de 2026-10-05 a 2026-10-09), em 1.292 requests, todos no Opus 5.5 (`docs/ai-costs.csv`). O valor é o equivalente em API a preço de tabela. Numa assinatura, a mesma redução se traduz em menos bloqueios por limite de uso.
  - As decisões mudam o fluxo de trabalho, as skills e as configurações do harness. Errar custa dinheiro a cada sessão, ou qualidade nas decisões do projeto, que é de aprendizado.
- Histórico:
  - 2026-10-09: criação (perguntas 1 a 7), com análise dos transcripts de `~/.claude/projects/` e dos evals do plugin caveman.
  - 2026-10-09: pergunta 4 completada com a medição do `/context` e a decisão de não desligar as skills sincronizadas do claude.ai; Saídas ajustada ao que saiu.
- Referências pesquisadas:
  - documentação do Claude Code: custos, boas práticas, configuração de modelo;
  - tabela de preços da API;
  - repositório do plugin caveman (benchmarks e evals).
  - Lista completa em [Fontes](#fontes).

Cada pergunta traz:

- como o mercado resolve;
- as opções comparadas;
- a recomendação para o projeto.

Recomendação não é decisão. As decisões saem (ver [Saídas](#saídas)):

- nas ADRs;
- no `CLAUDE.md`, nas skills e nas configurações do harness;
- nas issues de harness.

## Diagnóstico

### Por componente de preço

| Componente | Tokens | US$ | % |
|---|---:|---:|---:|
| Cache read: reler o contexto a cada request | 238,1 M | 47,63 | 48% |
| Cache write 1h: gravar no cache o que entrou no contexto | 2,57 M | 20,57 | 21% |
| Output: texto, raciocínio e entrada de ferramentas (ex.: conteúdo do `Write`) | 0,95 M | 19,04 | 19% |
| Overhead: chamadas que o transcript não registra | | 7,15 | 7% |
| Cache write 5m | 0,80 M | 3,99 | 4% |
| Input sem cache | 2 k | 0,01 | 0% |

Preços do Opus 5.5 por milhão de tokens: input US$ 4, output US$ 20, cache write 5m US$ 5, cache write 1h US$ 8, cache read US$ 0,20.

### Por tamanho do contexto

O tamanho médio do contexto foi de **184 k tokens por request**. Nenhuma sessão compactou: modelos com janela nativa de 1M só compactam perto de 967 k tokens.

| Contexto no request | Requests | US$ |
|---|---:|---:|
| até 50 k | 157 | 6,20 |
| 50 a 100 k | 243 | 9,48 |
| 100 a 150 k | 174 | 10,46 |
| 150 a 200 k | 181 | 12,70 |
| 200 a 250 k | 203 | 13,72 |
| acima de 250 k | 346 | **38,73** |

- Os requests acima de 250 k são 27% do total e respondem por 42% do custo.
- Sessões:
  - duram de 6 a 8 horas;
  - atravessam várias issues;
  - chegam a 459 k tokens de contexto.
- Teto teórico de economia em cache read, se o contexto nunca passasse de:
  - 200 k: US$ 11,32;
  - 150 k: US$ 17,69;
  - 100 k: US$ 25,91.

### Prefixo fixo

- O primeiro request de cada sessão principal já leva de 38 a 47 k tokens. Isso inclui:
  - system prompt e ferramentas;
  - instruções dos MCPs;
  - lista de skills;
  - `CLAUDE.md` com `conventions.md`;
  - memória;
  - plugin caveman.
- Um subagent começa com cerca de 8 k.
- Cada 1 k tokens de prefixo custa cerca de US$ 0,26 no período (1.292 requests × US$ 0,20/M).

### O que entra no contexto

| Origem | Caracteres | Exemplos |
|---|---:|---|
| WebFetch | 249 k | página de preços lida duas vezes (50 k cada) |
| `cat` de documentos | 206 k | `roadmap.md`, `journal.md`, `use-cases.md` e `discovery/ledger.md` inteiros |
| Outros Bash | 195 k | |
| `sed` | 143 k | |
| scripts Python de análise | 120 k | |
| `grep` | 92 k | |
| `gh` | 65 k | |

- Os documentos crescem a cada entrega. Hoje:
  - `discovery/ledger.md`: 75 KB;
  - `domain.md`: 19 KB;
  - `journal.md`: 17 KB;
  - `roadmap.md`: 16 KB.
- As skills mandam ler documentos inteiros na preparação.

### Turnos

- Dos 474 turnos do usuário, 179 (38%) têm até 20 caracteres ("sim", "s", "pode seguir").
- Cada turno relê o contexto inteiro pelo menos uma vez. Com o contexto médio, o piso é de cerca de US$ 0,04 por turno, ou ~US$ 7 só nas confirmações.

### Por issue e por tipo de trabalho

- "Sem issue" (refinamento, planejamento e harness na `main`): US$ 52,19 (53%).
- Issue #6 (discovery e ADRs do ledger): US$ 19,82.
- Subagents (`reviewer`, Explore): US$ 4,66 (5%).

## Resumo das recomendações

| # | Pergunta | Recomendação | Economia estimada |
|---|---|---|---|
| 1 | Gestão de contexto | uma sessão por etapa do fluxo, com o estado passado pelos documentos; o Claude sugere `/clear` ao fechar cada etapa; `autoCompactWindow` de 200 k como rede de segurança | 10 a 18% |
| 2 | Modelo | Opus como padrão; Sonnet nas sessões de tarefa delegada, escolhido no início da sessão | 5 a 10% |
| 3 | Esforço de raciocínio | manter o padrão `medium`; `effort: low` nas skills mecânicas | até 5%, sem medição |
| 4 | Prefixo fixo | medir com `/context`; desligar conectores e plugins sem uso neste projeto; não mexer no `CLAUDE.md` | 1 a 3% |
| 5 | Saídas de ferramenta | skills leem só a seção necessária; pesquisa web delegada a subagent | 3 a 8% |
| 6 | Caveman | manter, no nível `lite`; o efeito no custo é marginal; medir com A/B | ~2% |
| 7 | Fluxo | não pedir confirmação para o próximo passo já previsto no fluxo; perguntas fechadas agrupadas num só `AskUserQuestion` | 3 a 7% |

As economias não se somam: reduzir o contexto (1) diminui o ganho de todas as outras.

## 1. Gestão de contexto

**Pergunta.** Sessão longa, cobrindo várias issues, ou sessão curta por etapa? Como encerrar ou compactar, e por onde passa o estado de uma sessão para a próxima?

- Custo de errar: é o maior item do gasto (48% em cache read, 42% em requests acima de 250 k).
- Reversibilidade: total. É hábito e configuração, mudável a qualquer momento.

**Eliminatórios e critérios.**

- Eliminatórios:
  - não perder decisão nem contexto necessário para a tarefa em andamento;
  - manter o fluxo do `CLAUDE.md` (checkpoint, diário, revisão) funcionando.
- Critérios:
  - técnicos:
    - custo por request;
    - qualidade da resposta com contexto grande;
    - custo da troca de sessão.
  - do projeto: o dono precisa entender cada decisão, e a explicação acumulada na conversa tem valor.
  - de mercado: prática recomendada pela documentação oficial.

**Mercado.**

- Documentação oficial do Claude Code:
  - "Clear between tasks": `/clear` ao trocar de trabalho, porque contexto velho custa em toda mensagem seguinte.
  - Gasto alto inesperado "usually traces back to long sessions that were never cleared or to Opus left as the default model".
  - O desempenho cai com o contexto cheio: "LLM performance degrades as context fills". Isso é opinião do fornecedor, apoiada em estudo com medição sobre degradação em contexto longo (Chroma, *Context Rot*).
  - Escrever a spec em arquivo e "start a fresh session to execute it".
  - `/compact` lê a conversa inteira para resumir, então é um request grande. `/clear` não custa nada.
- Configuração disponível:
  - `autoCompactWindow`, de 100 k a 1M, compacta automaticamente ao atingir o limite;
  - `/compact <instruções>`;
  - "Summarize from here" no `/rewind`;
  - `/btw` para pergunta lateral que não entra no histórico.

**Opções.**

| Critério | A: status quo | B: sessão por etapa | C: B + autoCompact 200 k | D: compactação manual frequente |
|---|---|---|---|---|
| Custo por request | alto (média de 184 k) | baixo | baixo, com teto garantido | médio |
| Custo de troca | nenhum | ~US$ 0,32 (prefixo de 40 k no cache 1h) | B, mais ~US$ 0,50 por compactação | ~US$ 0,50 por compactação |
| Risco de perder contexto | nenhum | baixo: o estado está nos documentos | baixo | médio: o resumo perde detalhe |
| Depende de disciplina | não | sim | menos que B | sim |

- A: status quo.
  - Prós:
    - nenhuma explicação se perde;
    - nada a lembrar.
  - Contras:
    - 42% do custo em contexto acima de 250 k;
    - possível degradação de qualidade.
  - Limitações: o custo por turno cresce sem limite ao longo do dia.
- B: uma sessão por etapa do fluxo (refino, discovery, spec, implementação de uma tarefa), com `/clear` na troca.
  - Prós:
    - o projeto já grava o estado em arquivo (roadmap, discovery, spec, issue, diário);
    - é a prática recomendada;
    - custo zero.
  - Contras:
    - depende de lembrar;
    - a explicação dada no chat se perde se não foi para o documento.
  - Limitações: uma etapa longa sozinha (ex.: discovery do ledger) ainda passa de 200 k.
- C: B mais `autoCompactWindow: 200000` no `.claude/settings.json`.
  - Prós:
    - teto automático mesmo quando a disciplina falha;
    - cobre a etapa longa.
  - Contras:
    - a compactação custa e resume com perdas;
    - o momento da compactação não é escolhido.
  - Limitações: se compactar no meio de um raciocínio, pode exigir reler arquivos.
- D: manter sessões longas e rodar `/compact` com instruções a cada etapa.
  - Prós: preserva a continuidade da conversa.
  - Contras:
    - cada compactação é um request grande;
    - perde mais que um documento escrito.
  - Limitações: não resolve a disciplina, só a troca por outra.

**Custo-benefício.**

- Construir:
  - B: uma regra no `CLAUDE.md` (Fluxo por tarefa) e na `/refine`;
  - C: uma linha no settings.
- Operar: `/clear` e uma frase de retomada ("continue a issue #N").
- Reverter: apagar a linha.
- Benefício: de US$ 11 a 18 no período analisado (11 a 18%), menos as trocas (~US$ 0,32 cada).

**Riscos e limitações.** Pre-mortem: "um ano depois, deu errado porque..."

- Explicações dadas no chat se perderam com o `/clear`, e o dono deixou de entender decisões antigas. Mitigação: o checkpoint e o diário já registram o entendimento antes da troca de sessão.
- O limite de 200 k compactou no meio de uma revisão e o resumo perdeu achados. Mitigação: instrução de compactação no `CLAUDE.md` (preservar arquivos alterados, achados e decisões).

**Gatilho da troca.** O Claude não consegue rodar `/clear` sozinho: slash commands são do usuário, e hooks não encerram a sessão. O gatilho viável:

- o Claude sugere `/clear` ao fechar uma etapa do fluxo (critério na seção Sessões do `CLAUDE.md`);
- junto com a sugestão, ele dá a frase de retomada (ex.: "continue a issue #N, próximo passo: X");
- a retomada não depende da frase: um hook `SessionStart` (matcher `startup|clear`) prepara a sessão nova, e "continue" basta. Ele injeta:
  - a branch atual e a issue tirada do nome da branch;
  - os passos ainda abertos da issue;
  - o `git status` resumido.
- o hook não roda `/clear`, só prepara a sessão seguinte. Complemento acrescentado na issue #16, depois de uma sessão fechada sem querer no meio da issue, sem frase de retomada;
- a regra fica no `CLAUDE.md` (seção Sessões);
- opcional: status line com o uso do contexto, para o dono ver quando a sessão está cara.

**Recomendação:** C, com a sugestão de `/clear` pelo Claude no fim de cada etapa. A sessão por etapa resolve a maior parte do custo e casa com o fluxo, que já escreve tudo em arquivo. O `autoCompactWindow` de 200 k cobre os dias em que a disciplina falha.

**O que mudaria a recomendação:** se a qualidade cair visivelmente após compactações, subir o limite para 300 k ou ficar só com B.

## 2. Modelo por tipo de trabalho

**Pergunta.** Opus em tudo, ou modelo mais barato em parte do trabalho?

- Custo de errar: o Sonnet 5.5 custa metade do Opus 5.5 em todos os componentes. Modelo fraco em decisão de domínio custa retrabalho e entendimento errado.
- Reversibilidade: total.

**Eliminatórios e critérios.**

- Eliminatórios:
  - as tarefas do modo Manual e as decisões (discovery, ADR, spec, revisão) mantêm a qualidade atual.
- Critérios:
  - técnicos:
    - preço;
    - custo de troca de modelo (o cache é por modelo, então trocar no meio da sessão regrava o contexto todo).
  - do projeto: divisão do trabalho entre Manual, Pareado e Delegado.
  - de mercado: recomendação oficial.

**Mercado.**

- Documentação oficial: "Sonnet handles most coding tasks well and costs less than Opus. Reserve Opus for complex architectural decisions or multi-step reasoning". Para subagents simples, `model: haiku`.
- O alias `opusplan` usa Opus no plan mode e Sonnet na execução.
- `CLAUDE_CODE_SUBAGENT_MODEL` ou `model` no frontmatter do agent fixa o modelo do subagent.
- Não há benchmark público comparando Opus 5.5 e Sonnet 5.5 neste tipo de trabalho (documentação e decisão de arquitetura em português). A evidência é a recomendação do fornecedor.

**Opções.**

| Critério | A: Opus em tudo | B: modelo por sessão, segundo o modo | C: `opusplan` | D: Opus principal, subagents em Sonnet |
|---|---|---|---|---|
| Economia | 0 | 5 a 10% | 10 a 25% | ~2% (subagents são 5%) |
| Risco de qualidade | nenhum | baixo: Sonnet só no Delegado | médio: a execução do Pareado roda em Sonnet | baixo, mas atinge o `reviewer` |
| Troca de cache | não | não, se escolhido no início | sim, a cada entrada e saída do plan mode | não |

- A: Opus em tudo.
  - Prós: qualidade máxima, nada a decidir.
  - Contras: maior custo.
  - Limitações: paga Opus por boilerplate e infraestrutura.
- B: Opus como padrão; Sonnet nas sessões de tarefa delegada (infraestrutura, ferramentas, CI, boilerplate), escolhido com `/model` no início da sessão.
  - Prós:
    - casa com a divisão do trabalho do `CLAUDE.md`;
    - sem troca de cache no meio.
  - Contras:
    - exige lembrar de trocar;
    - a fatia delegada hoje é pequena (as issues #11 e #12 do `tokencost` somam US$ 12,52).
  - Limitações: a economia cresce só quando a fase de implementação tiver mais tarefa delegada.
- C: `opusplan`.
  - Prós: automático.
  - Contras:
    - o fluxo do projeto não usa plan mode;
    - o Pareado e o Manual (explicação, revisão) rodariam em Sonnet.
  - Limitações: não encaixa no fluxo atual.
- D: subagents em modelo menor.
  - Prós: simples (frontmatter).
  - Contras:
    - o `reviewer` é o portão de qualidade e custa pouco;
    - a economia é ~US$ 2.
  - Limitações: não ataca o grosso do custo.

**Custo-benefício.**

- Construir: uma linha na tabela "Divisão do trabalho" do `CLAUDE.md`.
- Operar: `/model sonnet` ao abrir sessão de tarefa delegada.
- Reverter: apagar a linha.
- Benefício: 50% do custo das sessões delegadas.

**Riscos e limitações.** Pre-mortem: o Sonnet errou em tarefa "delegada" que tocava regra de negócio, e o erro passou na revisão. Mitigação: a regra do `CLAUDE.md` já manda perguntar na dúvida sobre o modo; na dúvida, Opus.

**Recomendação:** B. Captura a economia onde a qualidade pesa menos, sem trocar o cache no meio da sessão. D não compensa: o ganho é pequeno e enfraquece o revisor.

**O que mudaria a recomendação:** benchmark ou experiência mostrando o Sonnet 5.5 equivalente em discussão de design. Nesse caso, avaliar Sonnet também no Pareado.

## 3. Esforço de raciocínio

**Pergunta.** Manter o esforço padrão em tudo, ou ajustar por tipo de trabalho?

- Custo de errar: o output é 19% do gasto, e só ~150 k dos 952 k tokens de output são texto visível. O resto é raciocínio e entrada de ferramentas.
- Reversibilidade: total.

**Eliminatórios e critérios.**

- Eliminatórios: não reduzir o raciocínio em decisões, revisões e tarefas do modo Manual.
- Critérios:
  - economia;
  - granularidade do controle;
  - risco de erro em tarefa mecânica.

**Mercado.**

- Documentação oficial:
  - o raciocínio é cobrado como output;
  - no Opus 5.5 não dá para desligar;
  - o padrão do Opus 5.5 já é `medium`;
  - o frontmatter de skill e de subagent aceita `effort`, que vale enquanto ele roda;
  - `effortLevel` no settings de usuário não se aplica ao Opus 5.5.
- Não há medição pública do ganho de `low` contra `medium` em tarefa mecânica. Nos transcripts deste projeto, o raciocínio fica oculto, então não dá para atribuir o gasto por skill.

**Opções.**

| Critério | A: padrão `medium` em tudo | B: `effort: low` nas skills mecânicas | C: `high` no Manual e nas decisões |
|---|---|---|---|
| Economia | 0 | pequena, não medida | negativa |
| Risco | nenhum | erro em passo mecânico com regra sutil (ex.: `/cards`) | nenhum |

- A:
  - Prós: nada a manter.
  - Contras: paga raciocínio médio em passo mecânico.
- B: `effort: low` em `/journal` e `/roadmap`, que só registram o que já foi decidido.
  - Prós:
    - granular;
    - reversível por skill.
  - Contras: ganho não medido.
  - Limitações: a `/cards` tem regras sutis (automações do quadro, testes do card) e não deve entrar.
- C: aumenta o custo; só se houver sinal de raciocínio raso nas decisões.

**Custo-benefício.** Construir e reverter custam uma linha por skill. O benefício é incerto: até 5% do total, se as skills mecânicas forem uma fatia relevante do output.

**Riscos e limitações.** Pre-mortem: o diário passou a omitir decisões porque a skill raciocinou menos sobre o que registrar. Mitigação: o diário passa pela revisão do `reviewer`.

**Recomendação:** A agora, e B como experimento em `/journal` e `/roadmap`. A evidência é fraca, e o padrão do Opus 5.5 já é `medium`.

**O que mudaria a recomendação:** uma forma de medir o raciocínio por skill (ex.: comparar o output do `tokencost` antes e depois).

## 4. Prefixo fixo

**Pergunta.** Enxugar o que vai em todo request?

- Custo de errar: baixo. Cada 1 k tokens de prefixo custa ~US$ 0,26 no período (0,3%).
- Reversibilidade: total.

**Eliminatórios e critérios.**

- Eliminatórios: manter as regras do `CLAUDE.md` que mudam o comportamento.
- Critérios:
  - tokens economizados;
  - perda de funcionalidade.

**Mercado.**

- Documentação oficial:
  - `CLAUDE.md` abaixo de 200 linhas, com instruções específicas movidas para skills;
  - desligar MCP sem uso com `/mcp`;
  - preferir CLI (`gh`) a MCP;
  - `/context` mostra o que ocupa o contexto.
- Situação deste projeto:
  - o `CLAUDE.md` com `conventions.md` tem ~8,7 KB (~2,5 k tokens), dentro do recomendado;
  - os conectores do claude.ai (Claude Docs, Google Drive) e as skills de plugins (anthropic-skills, caveman) entram em toda sessão, mesmo sem uso neste projeto.

**Opções.**

| Critério | A: nada | B: desligar o que não é usado aqui | C: B + reduzir o `CLAUDE.md` |
|---|---|---|---|
| Tokens economizados | 0 | estimados 3 a 6 k, a medir | B mais ~1 k |
| Perda | nenhuma | conectores fora deste projeto | regras do fluxo saem do contexto |

- A: nada muda; o prefixo é ~10% do cache read.
- B: rodar `/context`; desligar no projeto os conectores e plugins sem uso.
  - Prós: sem perda para o projeto.
  - Contras:
    - parte do prefixo (system prompt, ferramentas) não é controlável;
    - os conectores do claude.ai podem ser configurados por conta, não por projeto. A verificar.
- C: mover o fluxo do `CLAUDE.md` para skills.
  - Contras: o fluxo é usado em toda sessão, e movê-lo arrisca o Claude pular etapas.

**Custo-benefício.** B custa minutos e economiza 1 a 2%. C economiza ~0,3% e arrisca o fluxo.

**Riscos e limitações.** Pre-mortem: um conector desligado faz falta numa tarefa e ninguém lembra de religar. Mitigação: o próprio Claude avisa quando a ferramenta não está disponível.

**Recomendação:** B. Medir com `/context` antes de decidir o que desligar.

**O que mudaria a recomendação:** se o `/context` mostrar algo grande e inesperado, como a lista de skills ou a memória, tratar o item específico.

**Medição (2026-10-09).** O `/context` numa sessão nova, já com as mudanças da #16, mediu 31,1 k tokens de prefixo:

- prompt do sistema e ferramentas: 19,9 k, fora do controle do projeto;
- memória: 4,8 k (`CLAUDE.md` 3,3 k, `conventions.md` 1 k, `MEMORY.md` 0,4 k), acima dos ~2,5 k estimados, pelas regras novas desta issue;
- skills: 4,7 k, das quais ~2,1 k das skills sincronizadas do claude.ai (documentos, planilhas, apresentações), sem uso neste projeto;
- mensagens dos hooks: 1,5 k.

Os conectores do claude.ai não aparecem como MCP, e o `/plugin` não lista as skills sincronizadas: elas vêm da conta. Desligá-las valeria para todos os projetos, para economizar ~2 k de prefixo (~US$ 3 por mês). Decisão: não desligar.

## 5. Saídas de ferramenta

**Pergunta.** Limitar o que entra no contexto vindo de ferramentas?

- Custo de errar: médio. ~800 k caracteres de saída entraram no contexto e foram relidos em todos os requests seguintes da sessão. A pergunta 1 reduz esse efeito.
- Reversibilidade: total.

**Eliminatórios e critérios.**

- Eliminatórios: a skill continua lendo o que precisa para não contradizer documento existente.
- Critérios:
  - tokens evitados;
  - risco de ler pouco e errar;
  - complexidade.

**Mercado.**

- Documentação oficial:
  - hooks que pré-processam a saída (ex.: filtrar o `go test` para só as falhas);
  - delegar operações verbosas (documentação, logs, testes) a subagent, que devolve só o resumo;
  - skill com visão geral evita explorar arquivos.
- Neste projeto:
  - a maior saída foi WebFetch: a página de preços lida duas vezes, 50 k caracteres cada;
  - a segunda maior foi `cat` de documento inteiro (`roadmap.md`, `journal.md`, `discovery/ledger.md`), mandado pelas skills;
  - `go test` ainda não pesa (5 k caracteres).

**Opções.**

| Critério | A: nada | B: leitura por seção nas skills | C: subagent para pesquisa web | D: hook de filtro de saída |
|---|---|---|---|---|
| Tokens evitados | 0 | médio, crescente com os documentos | médio, concentrado em discovery | baixo hoje |
| Risco | nenhum | ler pouco e contradizer outra seção | o resumo omite detalhe | o filtro esconde informação útil |
| Construir | nada | ajustar as skills | uma linha na `/discovery` | script e settings |

- B: as skills leem o índice de títulos (`grep -n '^#'`) e só as seções relevantes; o `journal.md`, só as últimas entradas.
  - Prós: ataca o crescimento contínuo dos documentos.
  - Contras: exige que as seções tenham títulos claros.
- C: a `/discovery` delega a pesquisa web a um subagent, que devolve fatos com fonte.
  - Prós:
    - a página inteira fica fora do contexto principal;
    - o subagent pode rodar em Sonnet.
  - Contras: perde a leitura direta da fonte pelo agente principal.
- D: hook `PreToolUse` no Bash.
  - Contras: hoje não há saída de teste ou log grande para filtrar.
  - Limitações: passa a valer na fase de implementação.

**Custo-benefício.** B e C custam poucas linhas nas skills. Juntas economizam de 3 a 8%, e o ganho de B cresce com o tamanho dos documentos.

**Riscos e limitações.** Pre-mortem: a skill leu só uma seção do `domain.md` e escreveu uma regra que contradiz outra. Mitigação: o `reviewer` lê os documentos inteiros no contexto isolado dele.

**Recomendação:** B e C. D fica para quando a fase de implementação gerar saídas grandes de teste.

**O que mudaria a recomendação:** se a pergunta 1 deixar as sessões curtas, o ganho de B e C cai, porque a saída é relida por menos requests.

## 6. Caveman

**Pergunta.** O plugin caveman é efetivo? Manter, ajustar o nível ou remover?

- Custo de errar: baixo em dinheiro. O risco maior é a clareza das explicações, num projeto cujo objetivo é o dono dominar cada decisão.
- Reversibilidade: total.

**Eliminatórios e critérios.**

- Eliminatórios: as explicações do modo Manual e do checkpoint continuam compreensíveis.
- Critérios:
  - economia medida;
  - clareza das respostas;
  - custo do próprio plugin.

**Mercado.**

- O que o plugin afirma:
  - README: "~75% of output tokens";
  - tabela de benchmark: média de 65% (de 22 a 87%) contra resposta normal, sem grupo de controle.
- O próprio README admite: "Caveman only affects output tokens — thinking/reasoning tokens are untouched".
- Eval do repositório, com controle (`evals/snapshots/results.json`):
  - condições: Opus 4.6, 10 perguntas de uma só resposta, sem ferramentas;
  - mediana em caracteres: sem instrução 743; "Answer concisely." 826; caveman 362;
  - contra o controle "Answer concisely.", o caveman encurta ~56%;
  - a instrução genérica de concisão sozinha não encurtou nada.
  - Limitações do eval:
    - 10 perguntas;
    - tokenizador `o200k_base` (OpenAI), não o do Claude;
    - modelo anterior;
    - inglês;
    - sem ferramentas nem sessão longa.
- O artigo citado no README ("Brevity Constraints Reverse Performance Hierarchies", arXiv 2604.00025) mede respostas curtas em benchmark de acurácia, não em agente de código. Vale como indício, não como referência.
- Neste projeto:
  - o caveman esteve ativo em todas as sessões, então não há linha de base para comparar;
  - o texto visível é ~150 k dos 952 k tokens de output (~US$ 3 de US$ 19);
  - documentos, código e raciocínio, que são o grosso do output, não são afetados;
  - custo do plugin: ~700 tokens no início da sessão e ~45 por turno (lembrete do hook), menos de US$ 1 no período.
- Teto de efeito: se o caveman cortou 50% do texto visível, economizou ~US$ 3 de output, mais a releitura desse texto. Cerca de 2 a 4% do total. Remover custaria o mesmo tanto.

**Opções.**

| Critério | A: manter `full` | B: `lite` | C: remover |
|---|---|---|---|
| Economia | ~2 a 4% (estimada) | um pouco menor | 0 |
| Clareza | frases sem artigo nem conectivo, estranho em português | gramática completa, sem enrolação | normal |
| Explicações de conceito | risco de cortar o raciocínio que o dono precisa | menor risco | sem risco |

- A: manter `full`.
  - Prós: a maior economia do texto visível.
  - Contras:
    - prejudica a explicação, que é o centro do modo Manual e do checkpoint;
    - o estilo foi desenhado para inglês.
- B: `lite`.
  - Prós: corta a enrolação sem quebrar a frase.
  - Contras: economiza menos.
- C: remover.
  - Prós: menos um componente no harness.
  - Contras: o eval indica que a instrução genérica de concisão não substitui o plugin.

**Custo-benefício.** A diferença entre as opções é de poucos dólares no período. A decisão é de clareza, não de custo.

**Riscos e limitações.** Pre-mortem do `lite`: as respostas voltam a ficar longas, e o dono não percebe porque o efeito é pequeno. Mitigação: nenhuma necessária, já que o custo também é pequeno.

**Recomendação:** B. O caveman funciona no que mede (texto visível, ~56% contra um controle), mas o texto visível é uma fatia pequena do gasto. O nível `lite` mantém a clareza que o projeto de aprendizado exige. O efeito real é medido por um A/B com o `tokencost`: um período com `lite`, outro sem caveman, comparando o output por turno.

**O que mudaria a recomendação:**

- Um A/B medido com o `tokencost` (uma semana com, outra sem), se mostrar ganho acima de 5%.
- O dono preferir o estilo `full` na leitura.

## 7. Fluxo de trabalho

**Pergunta.** Que mudanças no fluxo (skills, regras de interação) reduzem o custo sem perder o controle do dono?

- Custo de errar: tirar pontos de decisão do dono contraria o objetivo do projeto. Manter o fluxo atual custa por turno.
- Reversibilidade: total.

**Eliminatórios e critérios.**

- Eliminatórios:
  - o dono continua decidindo cada ponto de decisão;
  - a regra "uma pergunta por vez" (memória do projeto) continua valendo para perguntas abertas e de discussão.
- Critérios:
  - turnos evitados;
  - controle do dono;
  - carga cognitiva das perguntas.

**Mercado.**

- A documentação oficial recomenda:
  - prompts específicos;
  - deixar o Claude rodar até um critério verificável;
  - corrigir cedo;
  - pré-aprovar ações seguras para reduzir interrupções.
- O `AskUserQuestion` aceita até 4 perguntas fechadas numa chamada, mostradas uma por vez e respondidas num só turno.
- Neste projeto:
  - 38% dos turnos são confirmações curtas;
  - várias seguem um "quer que eu faça X?" sobre o próximo passo já previsto no fluxo (criar a branch, gerar a ADR, seguir para o próximo passo);
  - cada uma relê o contexto inteiro.

**Opções.**

| Critério | A: status quo | B: sem confirmação do passo previsto | C: B + perguntas fechadas agrupadas |
|---|---|---|---|
| Turnos evitados | 0 | parte dos 38% | mais que B |
| Controle do dono | total | total nos pontos de decisão | total; responde várias de uma vez |
| Carga cognitiva | baixa | baixa | média, mas são perguntas fechadas |

- A: status quo.
  - Prós: nada muda.
  - Contras: piso de ~US$ 7 em confirmações, maior com sessões longas.
- B: regra no `CLAUDE.md`. O passo seguinte do fluxo documentado é executado sem pedir confirmação; a pergunta fica para decisões e para ações externas ou irreversíveis.
  - Prós:
    - menos turnos;
    - o fluxo já diz qual é o próximo passo.
  - Contras: o dono vê o resultado depois, não antes.
  - Limitações: não vale para o checkpoint, que só o dono aciona.
- C: B, e perguntas fechadas e independentes (ex.: nome da branch, label, sim/não) agrupadas num só `AskUserQuestion`.
  - Prós: um turno para várias respostas.
  - Contras: tensão com a regra "uma pergunta por vez", que nasceu para evitar blocos de texto.
  - Limitações: só para perguntas fechadas; discussão continua uma por vez.

**Custo-benefício.** B e C custam uma regra no `CLAUDE.md`. Economizam de 3 a 7%, e mais enquanto as sessões forem longas.

**Riscos e limitações.** Pre-mortem: o Claude tratou uma decisão como "passo previsto" e seguiu sem perguntar. Mitigação: a regra lista o que sempre exige pergunta:

- decisões de domínio e de arquitetura;
- push, PR, issue e qualquer ação externa;
- mudança de escopo.

**Recomendação:** B, e C só para perguntas fechadas. A regra "uma pergunta por vez" continua para discussão.

**O que mudaria a recomendação:** se o dono sentir perda de controle com B, voltar a confirmar os passos que geram artefato (ADR, spec).

## Saídas

| Saída | O quê |
|---|---|
| ADR | nenhuma: nada aqui é decisão de arquitetura do sistema; as regras ficam no `CLAUDE.md` e nas skills, e este discovery registra o porquê |
| `CLAUDE.md` | sessão por etapa, sugestão de `/clear` com frase de retomada no fim de cada etapa e instrução de compactação (1); modelo na tabela Divisão do trabalho (2); próximo passo sem confirmação e perguntas fechadas agrupadas (7) |
| `.claude/settings.json` | `autoCompactWindow: 200000` e hook `SessionStart` de retomada (1) |
| Skills | `effort: low` em `/journal` e `/roadmap` como experimento (3); leitura por seção nas skills (5); pesquisa web por subagent na `/discovery` (5) |
| Configuração local | medição com `/context`, sem desligar nada: as skills sincronizadas vêm da conta (4); caveman em `lite` (6) |
| Issue de harness | aplicar as saídas acima; a medição com o `tokencost` (contexto médio, custo por request, `effort: low`) e o A/B do caveman (6) ficam como testes da #16 depois do merge |
| `domain.md` | nenhuma: o tema é do harness, não do domínio |

## Fontes

- Referência: [Manage costs effectively (Claude Code)](https://code.claude.com/docs/en/costs)
- Referência: [Best practices for Claude Code](https://code.claude.com/docs/en/best-practices)
- Referência: [Model configuration (Claude Code): auto-compact, effort, opusplan](https://code.claude.com/docs/en/model-config)
- Referência: [Pricing (Claude API)](https://platform.claude.com/docs/en/about-claude/pricing)
- Referência: [Context Rot (Chroma Research)](https://research.trychroma.com/context-rot)
- Plugin analisado: [caveman (JuliusBrussee/caveman)](https://github.com/JuliusBrussee/caveman): README, `benchmarks/` e `evals/snapshots/results.json`
- Indício, não referência: [Brevity Constraints Reverse Performance Hierarchies in Language Models (arXiv 2604.00025)](https://arxiv.org/abs/2604.00025)
- Dados do projeto: `docs/ai-costs.csv` e transcripts em `~/.claude/projects/-home-betoth-go-src-github-com-betoth-cantina/`
