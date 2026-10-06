---
name: conventions
description: Cria ou atualiza docs/conventions.md (idiomas, linguagem, testes, dados, documentação, processo), em par com o usuário. Usar quando surgir uma convenção nova de código ou de documentação, ou quando uma existente precisar mudar.
argument-hint: <convenção, opcional>
---

Trabalhar nas convenções: **$ARGUMENTS**

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas.

## Preparação

1. Ler `docs/conventions.md` (se existir), `CLAUDE.md` e o código relacionado à convenção.
2. Se o arquivo não existir, criá-lo com as seções que já tiverem conteúdo real, entre: idiomas, linguagem (ex.: Go), testes, dados, documentação, processo. Garantir que `CLAUDE.md` o importa com `@docs/conventions.md`.

## Regras

- Convenção entra quando o primeiro código ou documento precisar dela, não antes.
- Só regras que valem para qualquer pessoa no projeto. Instruções específicas para o Claude ficam no `CLAUDE.md`.
- Uma linha por regra, no imperativo ou declarativa, sem justificativa longa. Se a justificativa for uma decisão significativa, ela vai para uma ADR (`/adr`) e a convenção aponta para ela.
- Convenção que muda código existente: avisar o usuário e listar os pontos afetados antes de registrar.

## Fechamento

- Mostrar só as linhas novas ou alteradas e pedir revisão.
