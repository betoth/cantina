---
name: refine
description: Refina um item antes de virar issue (entrega do roadmap, caso de uso, tarefa técnica, bug), em par com o usuário, e no fim oferece cadastrar a issue já pronta para começar. Usar quando o usuário quiser preparar, refinar ou detalhar um trabalho antes de executá-lo.
argument-hint: <ID do caso de uso, entrega do roadmap ou descrição>
---

Refinar: **$ARGUMENTS**

Nada vai para o GitHub antes de estar refinado. O roadmap é o backlog; a issue nasce pronta para começar (Ready).

## Forma de conduzir

- Uma pergunta por vez, com contexto curto e uma recomendação. Esperar a resposta antes da próxima.
- Mensagens curtas: no chat, só o trecho em discussão.

## Preparação

1. Se `$ARGUMENTS` estiver vazio, usar o item em discussão na conversa; se não estiver claro, perguntar qual e parar.
2. Ler `docs/roadmap.md`, o [template de issue](../cards/template.md) e os documentos ligados ao item.
3. Conferir se o item já tem issue (`gh issue list --search`). Se tiver, refinar e propor a edição do corpo em vez de criar outra.

## Caminho

- **Caso de uso:** conduzir a skill `/use-case` (se o detalhe não existir) e depois a `/spec` de cada spec listada no caso que ainda não estiver aprovada. Spec que depende de decisão cara de reverter, com alternativas reais: `/discovery` antes. A issue do caso de uso e as das tarefas das specs saem desses documentos.
- **Harness ou técnico:** refinar na conversa e montar o corpo da issue, com a seção Escopo preenchida.
- **Fora do roadmap** (parte de uma entrega, item que atravessa várias, bug): seguir "Issue fora do roadmap" da skill `/cards`. Se for algo que se pediria como entrega do produto, propor primeiro a entrada no roadmap (`/roadmap`).
- Pular o que já estiver pronto (caso detalhado, spec aprovada) e seguir do ponto em que parou.

## Refinamento de harness ou técnico

4. Levantar para si as lacunas: o que entra, o que não entra, como verificar, dependências, modo (manual, pareado, delegado, pela tabela do `CLAUDE.md`).
5. Perguntar cada lacuna ao usuário.
6. Decisão significativa (difícil de reverter, trade-off real): registrar com `/adr` e citar na issue. Com alternativas reais ainda não comparadas, `/discovery` antes da ADR.

## Ready

A issue está pronta quando tem:

- Escopo (ou spec aprovada nas Referências) com o que entra e o que não entra.
- Testes concretos, cada um automático ou manual, com o que observar.
- Pronto quando, citando o critério de pronto da fase que atende, se houver.
- Labels de tipo e de modo e milestone da fase.

## Fechamento

7. Mostrar o texto completo de cada issue e perguntar se pode cadastrar.
8. Com o ok, criar pela skill `/cards` (milestone se faltar, issue, quadro, link no roadmap).
