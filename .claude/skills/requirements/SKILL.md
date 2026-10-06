---
name: requirements
description: Cria ou atualiza os requisitos não funcionais em docs/non-functional-requirements.md, com IDs e metas testáveis, em par com o usuário. Usar ao definir ou revisar metas de carga, latência, privacidade, auditoria, disponibilidade, observabilidade ou retenção.
argument-hint: <grupo, opcional>
---

Trabalhar nos requisitos não funcionais. Grupo: **$ARGUMENTS**

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas. Nada de blocos de texto grandes: no chat, mostrar só o trecho em discussão; o texto completo fica no arquivo.
- Um grupo por vez.

## Preparação

1. Ler `docs/non-functional-requirements.md` (se existir), `docs/domain.md`, `docs/use-cases.md` e as ADRs.
2. Se o arquivo não existir, criá-lo com a introdução (metas declaradas e testáveis; regras de ID) e percorrer os grupos abaixo em ordem.

## Grupos

Cada grupo é uma seção com tabela `ID | Requisito`. Percorrer, propondo o mínimo que faz sentido para o projeto:

- **Carga e latência (`CARGA`):** partir de uma estimativa do uso real; apontar onde está o risco (volume ou concorrência).
- **Privacidade (`PRIV`):** quais dados pessoais existem, consentimento, acesso, isolamento, logs.
- **Auditoria (`AUD`):** o que é auditado, conteúdo do registro, imutabilidade, quem consulta.
- **Disponibilidade (`DISP`):** o que precisa continuar funcionando quando cada dependência cai.
- **Observabilidade (`OBS`):** traces, métricas técnicas e de negócio, logs, dashboards.
- **Retenção (`RET`):** quanto tempo cada tipo de dado fica.

## Regras

- ID no formato `RNF-GRUPO-NN`, estável, nunca reaproveitado.
- Todo requisito é verificável: valor numérico ou comportamento observável por teste. Sem meta que não possa ser medida.
- Limitação aceita (consequência conhecida que não será resolvida) é registrada como nota no grupo, sem ID.
- Requisito da v2 fica com prefixo "v2:" na descrição.
- Requisito que exige ação de um ator (ex.: aceitar consentimento) aponta para o caso de uso; se não existir, propor pela skill `/use-cases`.
- Requisito que muda uma regra do domínio também é registrado no `docs/domain.md`.

## Fechamento

- Mostrar só as linhas novas ou alteradas e pedir revisão.
- Lembrar que requisitos novos precisam entrar numa fase do roadmap (`/roadmap`).
