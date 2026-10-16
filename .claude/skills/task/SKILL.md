---
name: task
description: Conduz a execução de uma issue Ready, do branch ao PR (testes, implementação, verificação, checkpoint, diário e roadmap, revisão), e define quando ela está pronta.
argument-hint: <#issue>
disable-model-invocation: true
---

Executar a issue: **$ARGUMENTS**

Uma sessão por issue. O estado passa para a sessão seguinte pela issue (checkbox dos Passos e dos Testes) e pelo diário, não pela conversa.

## Preparação

1. Se `$ARGUMENTS` estiver vazio, usar a issue da branch atual (o hook de retomada a injeta); sem issue, perguntar qual e parar.
2. Ler a issue (`gh issue view`), e a mãe quando for sub-issue. Só issue Ready (critérios da skill `/refine`); se faltar algo, refinar antes.
3. Identificar o modo pela label e pela tabela do `CLAUDE.md`, e avisar se o modelo ativo não for o da tabela. Na dúvida sobre o modo, perguntar antes de escrever código.
4. Retomada: seguir do primeiro passo desmarcado da issue.

## Passos

1. **Branch** pela seção Git de `docs/conventions.md`; card em Em andamento (skill `/cards`).
2. **Ciclo por critério**, pelo papel do Claude no modo da issue, seguindo `docs/conventions.md` (lido aqui, já que o `CLAUDE.md` não o importa). Para cada critério de aceite, em ordem:
   - um teste no seam da issue ("Onde testar"), citando o `CA-NN` no nome ou comentário, visto falhando;
   - o código mínimo que o faz passar;
   - o próximo critério.
   Nada de escrever todos os testes antes do código: o teste de cada ciclo responde ao que o anterior ensinou. Mock só na fronteira com sistema externo.
3. **Verificação:** `make check` (build, lint, testes) verde. Enquanto o Makefile não tiver o alvo `check`, `go build ./... && go vet ./... && go test ./...` na raiz e `go -C tools build -o /dev/null ./... && go -C tools vet ./... && go -C tools test -count=1 ./...` para o módulo `tools/` (sem cache: há testes que leem arquivos do repo fora do pacote).
4. **Checkpoint de entendimento:** perguntar ao dono se quer rodar o `/checkpoint`. Só ele aciona, digitando `/checkpoint`: a skill não carrega pela ferramenta de skill.
5. **Diário e roadmap:** entrada do dia em `docs/journal.md`, pela skill `/journal` (com os pontos do checkpoint, se houver), e marcação no roadmap da entrega que a issue conclui, para irem no mesmo PR e passarem pela revisão.
6. **Revisão** pelo agent `reviewer`, sobre a branch, com o corpo da issue (e da mãe, se for sub-issue) na instrução, porque ele não tem `gh`.
   - Bloqueante: corrigir (pelo modo da issue) ou o dono descarta com motivo.
   - Correção em código segue o ciclo do passo 2: primeiro o teste que falha, depois o código.
   - Correção que muda o que o diário afirma atualiza também a entrada.
   - Sugestão não aplicada: o dono confirma se segue com ela pendente.
   - Depois de corrigir bloqueante, nova rodada; só bloqueante obriga outra rodada. Se ainda houver bloqueante após a terceira rodada, parar e pedir ao dono para decidir.
   - Mudança depois do PR aberto: nova rodada e seção `## Review` do PR atualizada, exceto quando muda só `docs/ai-costs.csv` e `docs/ai-costs.md` (custo de IA).
7. **Retro:** perguntar ao dono se quer rodar a `/retro`, sobretudo quando a revisão achou bloqueante ou a sessão travou em algum ponto. Só ele aciona, digitando `/retro`.
8. **PR**, testes do card e custo de IA, pela skill `/cards`.

Cada passo concluído: marcar o checkbox na issue (skill `/cards`).

## Pronto

- `make check` verde.
- Diário e roadmap atualizados.
- Revisão sem bloqueante pendente.
- Testes do card marcados, pela regra da skill `/cards`.

## Fechamento

Com o PR aberto, sugerir `/clear` e dar a frase de retomada (ex.: `/task #16`, com o próximo passo como informação). Antes, o que a próxima sessão precisa saber já está gravado (checkbox da issue, diário).
