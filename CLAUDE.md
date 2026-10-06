# Cantina Escolar

Backend de cantina escolar pré-paga em Go: carteira do aluno, recarga via Pix (fake), regras dos responsáveis, compra autorizada contra saldo e regras, ledger append-only em partida dobrada. Projeto de portfólio; o objetivo é o dono do repo dominar cada decisão e cada linha crítica.

- Domínio, glossário, invariantes do ledger e questões em aberto: [docs/domain.md](docs/domain.md)
- Casos de uso (IDs UC-ATOR-NN citados em specs e roadmap): [docs/use-cases.md](docs/use-cases.md)
- Requisitos não funcionais (IDs RNF-GRUPO-NN citados em specs e roadmap): [docs/non-functional-requirements.md](docs/non-functional-requirements.md)
- Decisões de arquitetura: [docs/adr/](docs/adr/)
- Specs de funcionalidades: `docs/specs/` (criada com a primeira spec)
- Diário: [docs/journal.md](docs/journal.md)
- Plano e status por fase: [docs/roadmap.md](docs/roadmap.md)
- Execução: issues no [quadro do GitHub Projects](https://github.com/users/betoth/projects/1), uma milestone por fase

## Arquitetura

Diagrama de containers: [docs/architecture.md](docs/architecture.md).

- Monorepo, um `go.mod`, um binário por serviço em `cmd/`.
- Portas e adaptadores com DDD tático leve: domínio puro, sem imports de infraestrutura; sem cerimônia estilo Java.
- Contextos: canteen (núcleo, contém os módulos ledger e identity), payments (integração com o Pix), notifications (genérico: só entrega mensagens).
- Ledger e identity: schema próprio, acessados só pela sua porta, sem FK para outros contextos. Ledger append-only: nada é editado ou apagado.
- Síncrono na transação: autorização de compra, cadastro, estorno. Todo o resto por evento (Kafka), publicado via outbox.
- Stack: PostgreSQL, sqlc + pgx, goose, Kafka (franz-go), OpenTelemetry, golangci-lint. Ver tabela no README.

## Convenções

@docs/conventions.md

## Fluxo por entrega

Do plano ao código, cada nível detalha o anterior:

1. **Entrega** no [roadmap](docs/roadmap.md), que é o backlog. O refinamento começa pela skill `/refine`, que conduz os passos seguintes.
2. **Caso de uso** detalhado em `docs/use-cases/`, pela skill `/use-case`, quando a entrega for um caso de uso. Lista as specs que o implementam.
3. **Spec** em `docs/specs/`, pela skill `/spec` (template em [.claude/skills/spec/template.md](.claude/skills/spec/template.md)): critérios de aceite numerados e testáveis, e as tarefas de implementação. Item de harness ou técnico não tem spec: o escopo fica no corpo da issue.
4. **Issues** criadas já prontas para começar, pela skill `/cards`, no [quadro](https://github.com/users/betoth/projects/1).

## Fluxo por tarefa

1. **Branch** pela seção Git de `docs/conventions.md`.
2. **Testes** derivados dos critérios de aceite que a tarefa cobre.
3. **Implementação.**
4. **Verificação:** `make check` (build, lint, testes) verde. Enquanto o Makefile não existir, `go build ./... && go vet ./... && go test ./...`.
5. **Diário:** entrada do dia em `docs/journal.md`, pela skill `/journal`, para ir no mesmo PR.
6. **Revisão** pelo agent `reviewer`, sobre a branch. Bloqueante: corrigir (pelo modo da tarefa) ou o dono descarta com motivo. Sugestão não aplicada: o dono confirma se segue com ela pendente.
7. **PR** com `Closes #N` e a seção Revisão; antes do último commit, marcar no roadmap a entrega que o PR conclui; testes do card (automáticos e manuais) marcados antes do merge.
8. **Checkpoint de entendimento:** ao concluir uma tarefa, perguntar ao dono se quer rodar o `/checkpoint`. Só ele aciona.

Pronto = `make check` verde + diário atualizado + revisão sem bloqueante pendente + testes do card marcados.

## Divisão do trabalho

| Modo | Tarefas | Papel do Claude |
|---|---|---|
| Manual | núcleo do domínio e regras de negócio; código cuja correção depende de concorrência, transação ou consistência entre sistemas; invariantes; primeira implementação de cada padrão novo | **Não escreve código de produção.** Testes são escritos em par. Explica conceitos, revisa, aponta bugs e faz perguntas. |
| Pareado | integrações com sistemas externos; novas instâncias de padrões já implementados manualmente; specs, ADRs e documentação | Propõe e escreve junto; o dono revisa e decide. |
| Delegado | boilerplate; infraestrutura e configuração (build, CI, containers); ferramentas de apoio (simuladores, script de dados de demo, dados de teste); dashboards | Faz, verifica e reporta. |

Na dúvida sobre o modo de uma tarefa, perguntar antes de escrever código.

## Harness

- Fluxos reutilizáveis do Claude ficam em skills (`.claude/skills/<nome>/SKILL.md`), não em commands (`.claude/commands/`). Command só quando tiver vantagem clara sobre skill, registrando o motivo.
- Por padrão, skills podem ser acionadas automaticamente pela `description`. Usar `disable-model-invocation: true` quando só o usuário deve decidir o momento (ex.: checkpoint).
- Agents (`.claude/agents/`) só quando o contexto isolado trouxer vantagem: olhar sem viés da conversa que produziu o trabalho, restrição de ferramentas (ex.: só leitura) ou paralelismo e volume. Fora isso, skill.

## Git

- Commit só quando pedido. Nunca push sem pedido.
