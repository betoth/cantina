# 0001. Migrations versionadas com goose

- Status: aceita
- Data: 2026-10-05

## Contexto e problema

O schema do banco vai evoluir junto com o código ao longo das fases, e parte das regras de negócio depende dele: a imutabilidade do ledger é garantida por permissões (`GRANT`/`REVOKE`), a idempotência por restrições `UNIQUE`, e cada contexto (canteen, ledger, payments) tem schema próprio. Esse schema precisa ser criado de forma idêntica no ambiente local, nos testes de integração (testcontainers) e no CI, e alterado sem perder dados.

## Fatores de decisão

- Schema versionado junto do código, com histórico de cada mudança.
- Mesmo caminho para criar o banco em desenvolvimento, testes e CI.
- Controle total do SQL, incluindo permissões, restrições e índices.
- Um conjunto de migrations e uma tabela de controle por schema, preservando o isolamento entre contextos.
- Migrations embutidas no binário (`//go:embed`).
- Comportamento previsível quando uma migration falha.

## Opções consideradas

Abordagem:

- Script SQL de inicialização do container (`docker-entrypoint-initdb.d`)
- Migrations versionadas com ferramenta dedicada

Ferramenta (se migrations versionadas):

- golang-migrate
- goose
- tern

## Decisão

Escolhida: **migrations versionadas com goose**.

Migrations versionadas porque o schema carrega regras de negócio (permissões do ledger, restrições de idempotência) e precisa evoluir sem recriar o banco. Script de inicialização só roda em banco vazio e não registra o que já foi aplicado.

goose porque usa um arquivo por migration, tem API de embed simples, permite tabela de controle por schema e não deixa o banco em estado "dirty" que exige intervenção manual após falha.

## Consequências

- Positivas: o mesmo conjunto de migrations cria o banco local, o dos testes e o do CI; mudanças de schema ficam revisáveis no histórico do git; permissões do ledger fazem parte do código.
- Negativas: cada mudança de schema exige uma migration nova; golang-migrate é um pouco mais adotado e tem mais material de referência.

## Prós e contras das opções

### Script de inicialização do container

- Prós: zero dependência.
- Contras: só roda com volume vazio; não evolui um banco existente; não registra o que foi aplicado; não serve aos testes fora do compose.

### golang-migrate

- Prós: a ferramenta mais usada; muitos drivers.
- Contras: dois arquivos por migration (`.up.sql` e `.down.sql`); falha no meio deixa estado "dirty" e exige `force` manual; não aceita migration em Go.

### goose

- Prós: um arquivo com `-- +goose Up` / `-- +goose Down`; API de embed limpa; aceita migration em Go; tabela de controle configurável por schema.
- Contras: ligeiramente menos popular que golang-migrate.

### tern

- Prós: do autor do pgx, bem integrado ao Postgres.
- Contras: nicho; pouco reconhecido.
