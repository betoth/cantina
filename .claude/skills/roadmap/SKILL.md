---
name: roadmap
description: Cria ou atualiza docs/roadmap.md (fases, entregas, requisitos por fase, critério de pronto) e verifica cobertura de casos de uso e requisitos, em par com o usuário. Usar ao planejar ou replanejar fases, ao concluir entregas, ou para checar se algo ficou sem fase.
argument-hint: <fase ou ação, opcional>
---

Trabalhar no roadmap: **$ARGUMENTS**

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas. Nada de blocos de texto grandes: no chat, mostrar uma fase por vez.

## Preparação

1. Ler `docs/roadmap.md` (se existir), `docs/use-cases.md`, `docs/non-functional-requirements.md`, `docs/domain.md` e as ADRs.
2. Se o arquivo não existir: propor primeiro só a estrutura macro (fases, tema, por que nessa ordem) e esperar aprovação; depois detalhar fase por fase.

## Estrutura

- Tabela de fases: número, tema, situação (rascunho, em revisão, planejada, em andamento, concluída).
- Por fase, só as seções que tiverem conteúdo: contexto (por que nessa posição), simplificações (o que é provisório e em que fase é substituído), casos de uso, requisitos não funcionais, técnico, ADRs, harness de IA, critério de pronto.
- Um checkbox por entrega verificável sozinha. Nada de agrupar várias entregas num item.
- Critério de pronto como demo ou teste observável.
- Requisito não funcional aparece só na fase que o entrega. Depois de entregue, vale para toda funcionalidade das fases seguintes, que precisa cumpri-lo sem item próprio no roadmap (ex.: auditoria entregue na Fase 1 vale para a recarga da Fase 2).

## Ordem das fases

- Primeiro o maior risco técnico, por último o mais padronizado.
- Respeitar dependências: um caso de uso não vem antes daquilo de que depende.
- O que for provisório numa fase (ex.: identidade fake, dados de seed) aparece em "simplificações" com a fase que o substitui.

## Checagem de cobertura

Ao criar, replanejar ou quando pedido, verificar e reportar uma lacuna por vez:

- Todo caso de uso do MVP está em exatamente uma fase.
- Todo requisito não funcional que não é v2 está em uma fase.
- Nenhum requisito é exigido antes da fase que o entrega (ex.: algo da Fase 1 que precisa de auditoria, entregue só na Fase 4).
- ADRs anotadas como pendentes estão em alguma fase.

## Atualização

- Item que entra numa fase já `em andamento`: marcar com `*(não planejado)*` depois do texto.
- Entrega concluída: marcar o checkbox. Fase com todas as entregas marcadas e critério de pronto atendido: situação `concluída`.

## Fechamento

- Mostrar só o que mudou e pedir revisão.
