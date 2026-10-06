---
name: domain
description: Cria ou atualiza docs/domain.md (visão, atores, glossário, contextos, regras de negócio, questões em aberto), em par com o usuário. Usar quando surgir ou mudar um conceito, regra ou ator do domínio, ou quando uma questão em aberto for respondida.
argument-hint: <tema, opcional>
---

Trabalhar no domínio. Tema: **$ARGUMENTS**

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas. Nada de blocos de texto grandes: no chat, mostrar só o trecho em discussão; o texto completo fica no arquivo.

## Preparação

1. Ler:
   - `docs/domain.md` (se existir);
   - `docs/use-cases.md`;
   - `docs/non-functional-requirements.md`;
   - `docs/conventions.md`;
   - as ADRs;
   - o discovery do tema em `docs/discovery/`, se existir.
2. Se `$ARGUMENTS` estiver vazio, usar o tema em discussão na conversa. Se o arquivo não existir, criá-lo seguindo a estrutura abaixo, começando pela visão e pelos atores.

## Estrutura de `docs/domain.md`

- **Visão:** um parágrafo sobre o que o sistema faz.
- **Atores:** tabela com ator, se faz login e papel (quem é, em uma linha). O que cada ator pode fazer fica só em `docs/use-cases.md`, com link.
- **Glossário:** termo, termo no código (em inglês, conforme convenções) e definição.
- **Contextos:** responsabilidade, eventos que publica e que consome; regras de fronteira.
- **Regras por tema:** uma seção por assunto do domínio (ex.: ledger, fluxo de compra, regras dos responsáveis). Regras declarativas, sem detalhe de implementação.
- **Escopo:** link para a lista de casos de uso e as capacidades técnicas do MVP que não são casos de uso.
- **Questões em aberto:** perguntas de domínio ainda sem decisão.

## Regras

- Sem duplicar outros documentos: o que cada ator faz fica nos casos de uso; metas de carga, privacidade etc. nos requisitos não funcionais; escolhas técnicas nas ADRs. Aqui, só link.
- Questão de domínio que surgir em qualquer trabalho vem para "Questões em aberto".
- Questão respondida: registrar a decisão na seção do tema e remover a questão da lista.
- Termo novo no domínio entra no glossário antes de aparecer em spec ou código.

## Fechamento

- Mostrar ao usuário só as seções novas ou alteradas e pedir revisão.
