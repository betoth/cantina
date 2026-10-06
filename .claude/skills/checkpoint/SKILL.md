---
name: checkpoint
description: Checkpoint de entendimento - faz perguntas sobre a tarefa recém-concluída para o usuário responder sem olhar o código, explicando na hora o que faltou.
argument-hint: <tarefa, opcional>
disable-model-invocation: true
---

Checkpoint de entendimento da tarefa: **$ARGUMENTS**

Objetivo: confirmar que o usuário domina as decisões e o código da tarefa: o porquê de cada escolha, o comportamento sob concorrência e falha, e as alternativas descartadas.

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas.
- O usuário responde sem olhar o código. Não mostrar código na pergunta.

## Preparação

1. Se `$ARGUMENTS` estiver vazio, usar a última tarefa concluída na conversa.
2. Ler o que mudou na tarefa (`git diff`, `git status`, arquivos citados na conversa), a spec e as ADRs relacionadas.
3. Escolher de 3 a 6 perguntas sobre o que é crítico na tarefa: concorrência, transação, invariantes, consistência entre sistemas, decisões de desenho. Nada de sintaxe ou nomes de função.

## Tipos de pergunta

Variar entre:

- **Por quê:** a razão de uma escolha ("por que X e não Y aqui?").
- **E se:** um cenário concreto ("duas requisições simultâneas com ...: o que acontece, passo a passo?").
- **Falha:** uma queda ou erro no meio do fluxo ("cai depois de A e antes de B: qual o efeito e o que protege?").
- **Mudança:** impacto de um requisito novo ("se a regra passasse a ser ..., o que mudaria?").

## Avaliação

4. Classificar cada resposta: certa, parcial ou errada.
5. Parcial ou errada: explicar na hora, de forma curta, o que faltou, apontando o arquivo e a linha que mostram. Depois seguir para a próxima pergunta.
6. Certa: confirmar em uma linha e seguir.

## Fechamento

7. Resumo curto: perguntas, resultado de cada uma e os pontos que faltaram.
8. Perguntar se o usuário quer registrar os pontos que faltaram em "Aprendizados" na entrada do diário (`/journal`).
