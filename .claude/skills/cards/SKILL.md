---
name: cards
description: Mantém as issues do projeto no GitHub (milestone por fase, issue refinada por entrega ou tarefa, quadro no Projects), em par com o usuário. Usar ao iniciar uma fase, ao cadastrar uma issue refinada, ao começar, abrir PR ou concluir uma entrega, quando o roadmap mudar, ou ao fechar uma fase.
argument-hint: <iniciar N | cadastrar | acompanhar | fechar N>
---

Cards do roadmap: **$ARGUMENTS**

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas: no chat, tabelas resumidas; o texto completo vai nas issues.
- Toda alteração no GitHub (criar, editar, mover card, marcar item, comentar, fechar): mostrar antes exatamente o que será feito e esperar o ok.

## Peças

| Peça | Regra |
|---|---|
| Milestone | uma por fase: `Fase N: tema` (tema da tabela do roadmap) |
| Issue | uma por checkbox de entrega da fase ou tarefa de spec, criada só depois de refinada (skill `/refine`), no formato de [template.md](template.md), em português |
| Quadro | Project `Cantina` (https://github.com/users/betoth/projects/1), público, ligado ao repositório, campo Status com A fazer, Em andamento, Em revisão, Feito |
| Automações do quadro | item adicionado → A fazer; PR ligado à issue → Em revisão; item fechado ou PR mergeado → Feito; item reaberto → Em andamento |
| PR | corpo com `Closes #N`; o merge fecha a issue |

## Preparação

1. Ler `docs/roadmap.md`, o [template](template.md) e a tabela de modos do `CLAUDE.md`.
2. `gh auth status`: precisa do escopo `project`. Sem ele, pedir ao usuário `! gh auth refresh -s project` e parar.
3. Se o quadro não existir (`gh project list --owner betoth`): propor criá-lo com o campo Status acima, ligado ao repositório. Automações e visão de quadro não têm API: passar ao usuário os passos na interface.

## Iniciar fase

O roadmap é o backlog: as issues não são criadas em lote, cada uma nasce do refinamento.

1. Mostrar a milestone que será criada e esperar o ok.
2. Criar a milestone; mudar a fase para `em andamento` pelas regras da skill `/roadmap`.
3. Sugerir a primeira entrega a refinar (`/refine`).

## Cadastrar

Chamado pela skill `/refine`, com o corpo da issue já aprovado pelo usuário.

1. Criar a milestone e as labels que faltarem.
2. Criar a issue (`gh issue create --body-file`) e adicioná-la ao quadro (a automação põe em A fazer).
3. Entrega do roadmap: acrescentar o link no checkbox: `- [ ] UC-OPER-01 Registrar compra ([#12](url))`.

## Acompanhar

- Início de uma entrega: só issue Ready (critérios da skill `/refine`); se faltar algo, refinar antes. Card para Em andamento; branch pela seção Git de `docs/conventions.md`.
- Passo da issue concluído: marcar o checkbox no corpo da issue.
- Antes do PR: revisão pelo agent `reviewer` (fluxo por tarefa do `CLAUDE.md`).
- Último commit antes do PR: marcar no roadmap o checkbox da entrega que o PR conclui (regras da skill `/roadmap`), para a marcação entrar no mesmo PR. Entrega com várias issues (caso de uso com tarefas): marcar no PR da última.
- PR aberto (formato pela seção Git de `docs/conventions.md`), corpo com `Closes #N` e a seção Revisão: cada achado e o que foi feito (corrigido, descartado com motivo, sugestão pendente com motivo); conferir que a automação moveu o card para Em revisão.
- Testes do card, antes do merge, na branch:
  - Automático: marcar quando passar no `make check` (ou no CI, quando existir).
  - Manual: o usuário executa o roteiro e informa o resultado; marcar o item e comentar na issue data, roteiro e resultado.
  - Falhou: comentar o que falhou na issue; card de volta para Em andamento.
  - Item de teste sem marcar: não seguir para o merge.
- PR mergeado: confirmar que a issue fechou e o card está em Feito.
- Entrega com issue que mudou de texto no roadmap: renomear a issue. Entrega removida: propor fechar a issue como não planejada.

## Issue fora do roadmap

Origem principal: a seção Tarefas de uma spec aprovada. Cada tarefa vira uma issue, com a spec nas Referências e os critérios de aceite que cobre na seção Testes.

- Parte de uma entrega: sub-issue da issue da entrega (`gh api graphql`, mutation `addSubIssue`), título com o ID da mãe: `[UC-OPER-01] Lançamento em partida dobrada`. A automação põe a sub-issue no quadro.
- Atravessa várias entregas: issue própria com o tipo no título (`[Técnico] Schema do ledger com migrations`), milestone da fase e, nas Referências, todas as entregas atendidas com link (`Relacionada a #12, #13`). Não entra no roadmap.
- Bug: como o caso anterior, título `[Bug] ...`, label `bug`.
- Se a issue for algo que se pediria como entrega do produto (não um meio para outra entrega), é entrega que faltou: entra primeiro no roadmap (skill `/roadmap`) e segue o refinamento.

## Fechar fase

1. Listar issues abertas da milestone. Se houver, reportar e parar.
2. Regressão na `main`: rodar de novo cada critério de pronto da fase (automáticos pelo `make check`, manuais pelo usuário). Critério que falhar: reportar e parar. O resultado vai na entrada de fechamento do diário.
3. Mostrar o que será feito (fechar a milestone) e esperar o ok.
4. Fechar a milestone; marcar a fase como `concluída` pelas regras da skill `/roadmap`.
5. Lembrar a entrada de fechamento no diário (`/journal`).
