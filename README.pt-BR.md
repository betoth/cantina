# Cantina Escolar

[English](README.md) · **Português**

Backend de cantina escolar pré-paga. Responsáveis recarregam a carteira do aluno via Pix e definem regras de consumo (bloqueio de produtos ou categorias, limite diário); o operador da cantina registra compras, autorizadas contra o saldo e as regras em uma única transação. Saldos são mantidos num ledger append-only em partida dobrada.

Sem front-end: a interface é a API HTTP (OpenAPI), uma CLI de caixa e simuladores.

## Documentação

- [Domínio](docs/domain.md): atores, glossário, contextos, ledger, fluxo de compra, regras
- [Arquitetura](docs/architecture.md): containers e como se comunicam
- [Casos de uso](docs/use-cases.md)
- [Requisitos não funcionais](docs/non-functional-requirements.md)
- [Decisões de arquitetura (ADRs)](docs/adr/)
- [Convenções](docs/conventions.md): idiomas, Go, testes, dados, processo
- [Roadmap](docs/roadmap.md): fases, entregas e critérios de pronto
- [Quadro](https://github.com/users/betoth/projects/1): andamento da fase atual

## Stack

| Ferramenta | Por quê |
|---|---|
| Go 1.27 | versão estável atual, fixada no `go.mod` |
| PostgreSQL | transações, `FOR UPDATE`, um schema por contexto |
| sqlc + pgx | SQL explícito com código tipado ([ADR 0002](docs/adr/0002-acesso-a-dados.md)) |
| goose | migrations versionadas em SQL, embutidas no binário ([ADR 0001](docs/adr/0001-migrations.md)) |
| Kafka (franz-go) | ordem por aluno, consumidores independentes, replay ([ADR 0003](docs/adr/0003-mensageria.md)) |
| Outbox | evento publicado só se a transação confirmar ([ADR 0004](docs/adr/0004-outbox.md)) |
| OpenTelemetry + Collector | instrumentação independente de fornecedor; backends (Jaeger, Prometheus) trocados pela configuração do Collector |
| golangci-lint | conjunto de linters padrão em Go; binário oficial |
| `tools/go.mod` | ferramentas Go versionadas sem misturar dependências com as da aplicação |
