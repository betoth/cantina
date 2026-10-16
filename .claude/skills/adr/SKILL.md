---
name: adr
description: Registra uma decisão de arquitetura em docs/adr no formato curto, em par com o usuário. Usar quando uma decisão técnica for difícil de reverter, surpreendente sem contexto e resultado de trade-off real (escolha de tecnologia, padrão, fronteira entre módulos, estratégia de consistência).
argument-hint: <título da decisão>
---

Registrar a decisão: **$ARGUMENTS**

ADRs são trabalho pareado: você propõe, o usuário revisa e decide.

## Forma de conduzir

- Mensagens curtas: no chat, só o trecho em discussão; o texto completo fica no arquivo.

## Preparação

1. Se `$ARGUMENTS` estiver vazio, usar a decisão em discussão na conversa; se não estiver clara, perguntar qual e parar.
2. Ler o [template](template.md) e as ADRs existentes.
3. A decisão merece ADR só quando é, ao mesmo tempo:
   - difícil de reverter;
   - surpreendente sem contexto (quem lê o código perguntaria "por que assim?");
   - resultado de um trade-off real, com alternativa séria.
   Se faltar um dos três, propor em vez da ADR: a seção Decisões da issue, ou uma linha na tabela Stack dos READMEs, e esperar o usuário decidir.
4. Se a decisão contradiz uma ADR aceita, ela substitui a anterior: avisar o usuário antes de seguir.
5. Alternativas ainda não comparadas, ou fato de mercado que falta: `/research` antes.

## Escrita

6. Próximo número: maior `NNNN` em `docs/adr/` + 1. Arquivo `NNNN-titulo-em-kebab-case.md`, em português. Status `proposta`, data de hoje.
7. Corpo de 1 a 3 frases: o contexto, a decisão e o porquê.
8. Seções opcionais do template só quando agregam:
   - Opções consideradas: só as discutidas de fato, com as fontes da `/research`; se faltar a alternativa óbvia que um revisor perguntaria, sugerir incluí-la;
   - Consequências: as negativas também, quando houver.
9. Ao substituir uma ADR: na antiga, mudar só o status para `substituída por [NNNN](NNNN-titulo.md)`; não editar o restante.
10. Se a decisão adiciona, remove ou muda a comunicação entre containers, atualizar o diagrama de `docs/architecture.md`.
11. Se a decisão envolve ferramenta listada na tabela Stack, atualizar `README.md` e `README.pt-BR.md` com o link para a ADR.

## Fechamento

12. Pedir revisão. Status passa a `aceita` só quando o usuário aprovar.
