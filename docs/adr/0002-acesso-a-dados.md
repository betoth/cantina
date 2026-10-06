# 0002. Acesso a dados com sqlc + pgx

- Status: aceita
- Data: 2026-10-05

## Contexto e problema

A correção do sistema depende de SQL preciso. A autorização de compra precisa, numa única transação, incrementar o gasto do dia com `UPDATE` condicional, debitar a carteira sem deixá-la negativa (`SELECT ... FOR UPDATE` ou `UPDATE ... WHERE saldo >= valor`) e garantir idempotência com `UNIQUE` e `ON CONFLICT`. Canteen e ledger precisam compartilhar essa transação. É preciso escolher como o código Go fala com o Postgres sem esconder esses detalhes.

## Fatores de decisão

- Locks, condições e restrições visíveis no código, revisáveis e testáveis.
- Controle explícito da transação, compartilhável entre módulos.
- Segurança de tipos entre SQL e Go.
- Pouco código repetitivo de mapeamento.
- Recursos específicos do Postgres (`FOR UPDATE`, `RETURNING`, `ON CONFLICT`, schemas).

## Opções consideradas

- ORM (ex.: GORM)
- pgx puro
- sqlc + pgx

## Decisão

Escolhida: **sqlc + pgx**. O SQL é escrito à mão, então locks e condições ficam explícitos; o sqlc gera código Go tipado e valida as queries contra o schema das migrations ([0001](0001-migrations.md)) na geração. As queries geradas recebem um `pgx.Tx`, o que permite que canteen e ledger participem da mesma transação.

## Consequências

- Positivas: o SQL crítico fica legível e revisável; erros de coluna ou tipo aparecem na geração, não em produção; menos código de `Scan`.
- Negativas: etapa de geração no fluxo (`sqlc generate`); queries muito dinâmicas exigem pgx direto.

## Prós e contras das opções

### ORM

- Prós: CRUD rápido; menos SQL para escrever.
- Contras: esconde locks e condições atrás de abstrações; recursos específicos do Postgres ficam em SQL bruto de qualquer forma; comportamento de transação menos explícito.

### pgx puro

- Prós: controle total; zero geração.
- Contras: mapeamento manual repetitivo; erros de SQL só aparecem em execução.

### sqlc + pgx

- Prós: SQL explícito; código tipado gerado; transação controlada pelo chamador.
- Contras: etapa de geração; menos flexível para SQL dinâmico.
