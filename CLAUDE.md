# Cantina Escolar

Backend de cantina escolar pré-paga em Go: carteira do aluno, recarga via Pix (fake), regras dos responsáveis, compra autorizada contra saldo e regras, ledger append-only em partida dobrada. Projeto de portfólio; o objetivo é o dono do repo dominar cada decisão e cada linha crítica.

- Domínio, glossário e invariantes do ledger: [docs/domain.md](docs/domain.md)
- Casos de uso (UC-ATOR-NN) e requisitos não funcionais (RNF-GRUPO-NN): [docs/use-cases.md](docs/use-cases.md), [docs/non-functional-requirements.md](docs/non-functional-requirements.md)
- Decisões de arquitetura: [docs/adr/](docs/adr/); convenções: [docs/conventions.md](docs/conventions.md); harness: [docs/harness.md](docs/harness.md)
- Plano e diário: [docs/roadmap.md](docs/roadmap.md), [docs/journal.md](docs/journal.md); execução no [quadro](https://github.com/users/betoth/projects/1)
- Fluxo: refino pela skill `/refine` (issue Ready, numa sessão), execução pela `/task` (uma sessão por issue)

## Arquitetura

- Monorepo, um `go.mod`, um binário por serviço em `cmd/`. Containers: [docs/architecture.md](docs/architecture.md).
- Portas e adaptadores com DDD tático leve: domínio puro, sem imports de infraestrutura; sem cerimônia estilo Java.
- Contextos: canteen (núcleo, com os módulos ledger e identity), payments (Pix), notifications (só entrega).
- Ledger e identity: schema próprio, acesso só pela porta, sem FK para outros contextos. Ledger append-only.
- Síncrono na transação: autorização de compra, cadastro, estorno. O resto por evento (Kafka), via outbox.

## Divisão do trabalho

| Modo | Tarefas | Papel do Claude | Modelo |
|---|---|---|---|
| Manual | núcleo do domínio e regras de negócio; código cuja correção depende de concorrência, transação ou consistência entre sistemas; invariantes; primeira implementação de cada padrão novo | **Não escreve código de produção.** Testes são escritos em par. Explica conceitos, revisa, aponta bugs e faz perguntas. | Opus |
| Pareado | integrações com sistemas externos; novas instâncias de padrões já implementados manualmente; issues, ADRs e documentação | Propõe e escreve junto; o dono revisa e decide. | Opus |
| Delegado | boilerplate; infraestrutura e configuração (build, CI, containers); ferramentas de apoio (simuladores, dados de demo e de teste); dashboards | Faz, verifica e reporta. | Sonnet |

Na dúvida sobre o modo, perguntar antes de escrever código. O modelo é escolhido no início da sessão e não muda no meio (a troca descarta o cache); na dúvida, Opus.

## Interação

- Sempre perguntar antes de:
  - decisão de domínio ou de arquitetura;
  - ação externa ou irreversível (push, PR, issue, card);
  - mudança de escopo;
  - checkpoint.
- O passo seguinte do fluxo documentado é executado sem pedir confirmação.
- Perguntas fechadas e independentes juntas num só `AskUserQuestion`; abertas, uma por vez.

## Git

- Commit só quando pedido. Nunca push sem pedido.

## Compactação

Preservar:

- arquivos alterados;
- achados da revisão e o que foi feito com cada um;
- decisões e o motivo;
- issue e passo em andamento.
