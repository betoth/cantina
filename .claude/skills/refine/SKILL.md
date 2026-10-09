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

Documentos de `docs/` são lidos por seção: primeiro o índice de títulos (`grep -n '^#' <arquivo>`), depois só as seções ligadas ao tema. O documento inteiro só quando a tarefa o percorre todo (ex.: checagem de cobertura).

1. Se `$ARGUMENTS` estiver vazio, usar o item em discussão na conversa; se não estiver claro, perguntar qual e parar.
2. Ler `docs/roadmap.md`, o [template de issue](../cards/template.md) e os documentos ligados ao item.
3. Conferir se o item já tem issue (`gh issue list --search`). Se tiver, refinar e propor a edição do corpo em vez de criar outra.

## Caminho

- **Caso de uso:** conduzir uma etapa por sessão (seção Sessões do `CLAUDE.md`), a próxima que faltar:
  - `/use-case`, se o detalhe não existir;
  - `/discovery` e as saídas dele (`/adr`, `/domain`), se a próxima spec depender de decisão cara de reverter, com alternativas reais;
  - `/spec` de cada spec listada no caso que ainda não estiver aprovada.
  - As issues saem desses documentos e são cadastradas pelo Fechamento só quando a última spec do caso for aprovada: primeiro a issue do caso de uso, depois as das tarefas, como sub-issues dela.
  - No fim de cada etapa, sugerir `/clear` e dar a frase de retomada, que nomeia o refino de origem (ex.: "continue o refino do UC-OPER-01, próximo passo: spec da compra").
- **Harness ou técnico:** refinar na conversa e montar o corpo da issue, com a seção Escopo preenchida.
- **Fora do roadmap** (parte de uma entrega, item que atravessa várias, bug): seguir "Issue fora do roadmap" da skill `/cards`. Se for algo que se pediria como entrega do produto, propor primeiro a entrada no roadmap (`/roadmap`).
- Refino com etapas em várias sessões: a sessão nova chama a `/refine` de novo, que pula o que já estiver pronto (caso detalhado, discovery e ADR gravados, spec aprovada) e segue do ponto em que parou.
  - Todas as specs aprovadas e issues faltando (a sessão caiu antes do Fechamento): ir direto ao Fechamento, conferindo com `gh issue list --search` quais já existem.

## Refinamento de harness ou técnico

4. Levantar para si as lacunas: o que entra, o que não entra, como verificar, dependências, modo (manual, pareado, delegado, pela tabela do `CLAUDE.md`).
5. Perguntar cada lacuna ao usuário.
6. Decisão significativa (difícil de reverter, trade-off real): registrar com `/adr` e citar na issue. Com alternativas reais ainda não comparadas, `/discovery` antes da ADR. Discovery e ADR são etapas próprias: pausar o refino gravando o que já foi decidido e as lacunas restantes numa linha da entrada do dia em `docs/journal.md`, e sugerir `/clear` com a frase de retomada, que nomeia o refino de origem (ex.: "próximo passo: ADR Y, depois continue o refino de X"). Depois das saídas do discovery (ou da ADR), a `/refine` chamada de novo retoma das lacunas gravadas e segue para o Ready e o Fechamento.

## Ready

A issue está pronta quando tem:

- Escopo (ou spec aprovada nas Referências) com o que entra e o que não entra.
- Testes concretos, cada um automático ou manual, com o que observar.
  - Automático por padrão.
  - Manual só com ganho sobre o automático: o comportamento depende de algo que o teste não alcança (render no GitHub, serviço externo real, automação do quadro) ou exige o olho do dono.
  - Eficientes:
    - cada teste cobre um critério que nenhum outro cobre;
    - o manual não repete o que um automático já verifica;
    - o roteiro manual é curto, com o comando exato e o que observar.
  - Manual que só se observa depois do merge: marcado `(manual, depois do merge)`.
- Pronto quando, citando o critério de pronto da fase que atende, se houver.
- Labels de tipo e de modo e milestone da fase.

## Fechamento

7. Mostrar o texto completo de cada issue e perguntar se pode cadastrar.
8. Com o ok, criar pela skill `/cards` (milestone se faltar, issue, quadro, link no roadmap).
