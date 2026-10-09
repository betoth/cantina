---
name: spec
description: Cria uma spec de funcionalidade em docs/specs a partir do template, em par com o usuário. Usar quando o usuário quiser especificar, detalhar ou definir os critérios de aceite de uma funcionalidade antes de implementá-la.
argument-hint: <nome da funcionalidade>
---

Criar a spec da funcionalidade: **$ARGUMENTS**

Specs são trabalho pareado: você propõe, o usuário revisa e decide.

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas. Nada de blocos de texto grandes: no chat, mostrar só o trecho em discussão; o texto completo fica no arquivo.

## Preparação

Documentos de `docs/` são lidos por seção: primeiro o índice de títulos (`grep -n '^#' <arquivo>`), depois só as seções ligadas ao tema. O documento inteiro só quando a tarefa o percorre todo (ex.: checagem de cobertura).

1. Ler:
   - o template [template.md](template.md);
   - `docs/use-cases.md` e o detalhe dos casos de uso envolvidos em `docs/use-cases/`, se existir;
   - `docs/non-functional-requirements.md`;
   - `docs/domain.md`;
   - `docs/conventions.md`;
   - as ADRs em `docs/adr/` relacionadas à funcionalidade;
   - o discovery do tema em `docs/discovery/`, se existir.
2. Ler specs existentes em `docs/specs/` que tenham relação com esta, para não contradizê-las.
3. Se `$ARGUMENTS` estiver vazio, usar a funcionalidade em discussão na conversa; se não estiver clara, perguntar qual e parar.

## Perguntas

4. Listar para si as lacunas que impedem escrever critérios de aceite testáveis: questões em aberto do `domain.md` que afetam esta funcionalidade, casos de erro sem comportamento definido, regras ambíguas.
5. Perguntar ao usuário sobre cada lacuna, com contexto curto e uma recomendação.
6. Tamanho: a spec entrega valor observável pelo ator (ou é base reaproveitada por várias, como o ledger) em até 6 tarefas. Se passar disso, propor a divisão antes de escrever.

## Escrita

7. Próximo número: maior `NNNN` em `docs/specs/` + 1 (a primeira é 0001; criar a pasta se não existir). Nome do arquivo: `NNNN-nome-em-kebab-case.md`, em português.
8. Preencher todas as seções do template, incluindo os casos de uso que a spec implementa e os requisitos não funcionais que ela precisa cumprir: os da fase atual e todos os já entregues em fases anteriores que se apliquem (consultar o roadmap). Status `rascunho`, data de hoje.
9. Fluxo: fluxograma Mermaid obrigatório, escrito antes dos critérios de aceite, com todas as decisões e caminhos de erro; cada caminho de erro tem linha correspondente na tabela de erros. Sequência: diagrama Mermaid obrigatório quando o fluxo atravessa mais de um serviço; omitir a seção caso contrário.
10. Critérios de aceite: numerados, no formato dado/quando/então, cada um verificável por um teste. Incluir a subseção de concorrência sempre que houver estado compartilhado.
11. Para cada requisito não funcional listado na spec, ao menos um critério de aceite que o verifique nesta funcionalidade (ex.: "então existe registro de auditoria com ..."). Requisito entregue antes não é presumido: é verificado de novo aqui.
12. Invariantes do ledger que a funcionalidade toca: citar pelo número do `domain.md`, sem reescrever.
13. Tarefas: quebrar a implementação em tarefas na ordem em que serão feitas, do domínio para fora (domínio puro, persistência, porta, API). Cada tarefa verificável sozinha, citando os critérios de aceite que cobre; todo critério coberto por ao menos uma tarefa.
14. Lacunas que o usuário decidir adiar vão para "Questões em aberto" da spec.

## Fechamento

15. Questões do `domain.md` resolvidas durante a spec: registrar a decisão na seção correspondente do `domain.md` e remover a questão da lista de abertas.
16. Mostrar ao usuário um resumo dos critérios de aceite e das tarefas e pedir revisão. Status passa a `aprovada` só quando o usuário aprovar. Com a spec aprovada, as tarefas viram issues pela skill `/cards`; num caso de uso, pelo Fechamento da `/refine`, quando a última spec do caso for aprovada.

Não escrever testes nem código neste fluxo.
