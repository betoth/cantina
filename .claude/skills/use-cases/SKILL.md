---
name: use-cases
description: Mantém a lista de casos de uso em docs/use-cases.md (adicionar, ajustar, descartar, mover entre MVP e v2) e revisa lacunas, em par com o usuário. Usar quando surgir um caso de uso novo, quando um mudar de escopo, ou para revisar se falta algum.
argument-hint: <ação ou caso, opcional>
---

Trabalhar na lista de casos de uso: **$ARGUMENTS**

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas. Nada de blocos de texto grandes: no chat, mostrar só o trecho em discussão; o texto completo fica no arquivo.

## Preparação

Documentos de `docs/` são lidos por seção: primeiro o índice de títulos (`grep -n '^#' <arquivo>`), depois só as seções ligadas ao tema. O documento inteiro só quando a tarefa o percorre todo (ex.: checagem de cobertura).

1. Ler `docs/use-cases.md`, `docs/domain.md` e `docs/non-functional-requirements.md`.
2. Se `docs/use-cases.md` não existir, criá-lo com: definição de caso de uso, onde fica o detalhe (`docs/use-cases/UC-ATOR-NN-nome.md`) e as specs, as regras abaixo, e as seções `## MVP` e `## v2`, cada uma com uma tabela `ID | Caso de uso` por ator.

## O que é caso de uso

- Objetivo de um ator, de ponta a ponta, do ponto de vista dele.
- Não é caso de uso: etapa interna de outro caso (ex.: webhook dentro da recarga), infraestrutura (ex.: outbox), regra de permissão (vai para o domínio), regra de negócio (vai para o domínio ou para a spec).
- O que é comum a todo ator com login (ativar conta, recuperar senha, autenticar) fica no grupo `USR`.

## Regras de ID

- Formato `UC-ATOR-NN`, prefixo do ator e número sequencial dentro do grupo.
- Caso novo recebe o próximo número do grupo.
- Objetivo mudou: caso novo, ID novo.
- Removido: vai para a seção "Descartados" (criada no primeiro); ID nunca reaproveitado.
- Mudança de escopo (v2 → MVP) move a linha de seção sem mudar o ID.
- Renumerar só é permitido enquanto nenhum documento cita o ID; conferir com busca antes.
- O número não define ordem de implementação; a ordem vem do roadmap.

## Revisão de lacunas

Quando pedida, passar ator por ator, uma lacuna por vez, com recomendação de escopo (MVP, v2 ou questão do domínio):

- Ciclo de vida de cada ator: entrada (cadastro, ativação, consentimento), uso diário, exceções (erro, perda, bloqueio), saída (desativação).
- Para cada entidade do domínio: quem cria, quem altera, quem desativa.
- Propriedades do domínio que nenhum caso usa (ex.: estado "bloqueada" sem caso que bloqueie).
- Requisitos não funcionais que exigem ação de algum ator (ex.: consentimento, auditoria).
- O mesmo dado alterado por atores diferentes com permissões diferentes vira casos separados.

## Fechamento

- Questões de domínio que surgirem vão para `docs/domain.md`.
- Mostrar só as linhas novas ou alteradas e pedir revisão.
