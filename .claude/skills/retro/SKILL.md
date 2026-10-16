---
name: retro
description: Retrospectiva de uma sessão de trabalho: acha onde o Claude travou, errou ou o reviewer apontou problema, e propõe a mudança no harness (check automático, regra do reviewer, ponteiro, corte no CLAUDE.md) que evita a repetição.
argument-hint: <sessão, opcional; padrão: a atual>
disable-model-invocation: true
---

Retrospectiva: **$ARGUMENTS**

A retro muda o ambiente, não o código: para cada problema, pergunta o que no harness deixou ele acontecer. Só propõe; nada muda até o dono escolher.

## Preparação

1. Fonte: a sessão atual, salvo se `$ARGUMENTS` apontar outra (transcript em `~/.claude/projects/`). Sessão longa já fora da janela: sugerir `/clear` e rodar a `/retro` apontando o transcript anterior.
2. Ler `docs/harness.md`, o relatório do `reviewer` da sessão, se houver, e os checks existentes (alvo `check` do Makefile, testes em `tools/`, CI).

## Candidatos

3. Levantar candidatos só a partir de momentos concretos da sessão; candidato que não aponta um momento é descartado. Por categoria, e para onde vai a correção:

| O que aconteceu | Correção |
|---|---|
| Demorou a achar um arquivo ou fato | ponteiro num arquivo que já é lido |
| Erro que uma ferramenta pegaria | check automático (teste em `tools/`, lint, passo do CI); check que existe mas não roda: ligá-lo, não criar outro |
| Achado do `reviewer` que é julgamento | regra que o `reviewer` cobra (`docs/conventions.md` ou a skill do artefato) |
| Erro que o `reviewer` deixou passar | regra nova ou mais clara para o `reviewer` |
| `CLAUDE.md` com instrução que não serve a toda sessão | mover para skill, convenção ou check |
| Chamada cara para o que devolveu (ver `tokencost`) | simplificar a ferramenta ou a leitura |
| Instrução que não muda o comportamento | apagar (no-op) |
| Informação que o Claude não conseguiu acessar | ampliar o acesso (log, comando, leitura de serviço) |

4. Antes de propor regra, classificar a violação, pelo princípio de `docs/harness.md`:
   - mecânica (padrão fixo, arquivo fora do lugar, link quebrado): check automático;
   - julgamento: regra para o `reviewer`, não para quem implementa;
   - `CLAUDE.md` só quando vale para toda sessão.
5. Achado do `reviewer` corrigido na sessão: apontar a regra que o teria evitado, ou dizer que nenhuma faltou.

## Fechamento

6. Mostrar os candidatos, do mais grave ao menos, cada um com o momento da sessão, a categoria e a correção, numerados para o dono escolher.
7. Aplicar só os escolhidos, pelo modo da issue, na branch:
   - issue de harness: a dela;
   - outra issue (produto, técnico, refino): uma própria a partir da `main`, `feat/<N>-harness-<slug>` com `N` da issue da sessão, para revisar e reverter o harness sem a entrega. Correção que depende da entrega fica na branch da issue.
8. Registrar no diário (Aprendizados e revisões), na branch da issue, cada correção aplicada: o achado, a causa no harness e a regra ou o check que entrou. Depois, nova rodada do `reviewer` sobre a mudança.
9. PR da branch própria: título `feat(harness): ...`, referência à issue pela seção Git de `docs/conventions.md` e o mesmo registro no `## Review`. Na branch da issue, o registro vai no `## Review` do PR dela (passo PR da `/task`).
