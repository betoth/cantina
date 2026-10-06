---
name: use-case
description: Detalha um caso de uso em docs/use-cases/UC-ATOR-NN-nome.md (fluxo principal, fluxos alternativos, resultado, specs), em par com o usuário. Usar quando for preciso descrever de ponta a ponta o que um caso de uso faz, antes de escrever as specs que o implementam.
argument-hint: <ID do caso de uso>
---

Detalhar o caso de uso: **$ARGUMENTS**

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas. Nada de blocos de texto grandes: no chat, mostrar só o trecho em discussão; o texto completo fica no arquivo.

## Preparação

1. Se `$ARGUMENTS` estiver vazio, usar o caso em discussão na conversa; se não estiver claro, perguntar qual e parar.
2. Ler a linha do caso em `docs/use-cases.md`, `docs/domain.md`, `docs/non-functional-requirements.md` e as specs que já citam o caso.
3. Ler o template [template.md](template.md).

## Escrita

4. Arquivo `docs/use-cases/UC-ATOR-NN-nome-em-kebab-case.md`, em português. Criar a pasta se não existir.
5. Preencher o template seção por seção, mostrando cada uma e esperando confirmação antes da próxima.
6. Sem detalhe técnico: o quê e o porquê, na visão do ator. Endpoints, tabelas, locks e eventos ficam nas specs.
7. Na tabela de specs, spec ainda não criada aparece só com número e nome, sem link; o link entra quando a spec for criada.
8. Regras de negócio: citar a seção do `domain.md`, sem reescrever. Regra nova vai para o `domain.md`.
9. Requisitos não funcionais: citar pelos IDs.

## Fechamento

10. Em `docs/use-cases.md`, transformar o ID do caso em link para o arquivo.
11. Questões que surgirem: de domínio vão para `docs/domain.md`; as do próprio caso ficam na seção do arquivo.
12. Pedir revisão final.
