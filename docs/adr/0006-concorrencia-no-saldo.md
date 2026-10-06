# 0006. Concorrência no saldo

- Status: aceita
- Data: 2026-10-06

## Contexto e problema

No recreio, dois caixas podem debitar a mesma carteira no mesmo instante, e toda compra de uma cantina credita a mesma conta de receita. Se duas transações lerem o mesmo saldo e gravarem por cima uma da outra, a carteira fica negativa ou um débito some, e no ledger append-only o erro não se apaga, só se compensa. É preciso escolher:

- como o saldo é conferido e atualizado sob concorrência;
- como tratar as contas que entram em quase toda transação;
- o que acontece quando uma transação falha por motivo transitório.

A análise completa está no [discovery do ledger](../discovery/ledger.md), perguntas 2, 3 e 12.

## Fatores de decisão

- Eliminatórios:
  - Invariante 4 do `domain.md`: conta que não pode ficar negativa nunca fica.
  - Bloqueio de carteira (`domain.md`): carteira bloqueada não é debitada; créditos continuam entrando.
  - RNF-CARGA-03: 20 compras simultâneas na mesma carteira; todas respondem, aprovadas ou recusadas, sem erro.
  - RNF-CARGA-01 e RNF-CARGA-02: 100 compras/s com p99 < 200 ms.
  - RNF-AUD-06 e [ADR 0004](0004-outbox.md): débito, compra, auditoria e outbox na mesma transação do Postgres.
  - [ADR 0005](0005-modelo-contabil-do-ledger.md): saldo resultante e versão em cada lançamento.
- Técnicos:
  - corretude evidente no código;
  - nenhuma tentativa desperdiçada na disputa;
  - sem deadlock;
  - latência previsível.
- Domínio: recusa por saldo, limite ou bloqueio é resultado de negócio, não erro.
- Mercado: técnica usada por ledgers de referência (Modern Treasury, TigerBeetle, pgledger) e prática documentada pelo Postgres.

## Opções consideradas

Técnica de concorrência no saldo:

- Update condicional em `READ COMMITTED`
- `SELECT ... FOR UPDATE` em ordem de ID e checagem em Go
- Versão otimista
- `SERIALIZABLE` com repetição

Eliminadas antes da comparação, por ficarem fora da transação do Postgres:

- lock distribuído (ex.: Redlock);
- fila com escritor único por conta.

Contas quentes (discovery, pergunta 3):

- tudo materializado e síncrono;
- saldo da conta quente aplicado de forma assíncrona, como no Modern Treasury;
- conta quente sem saldo materializado;
- subcontas.

Recusa depois de escrever (discovery, pergunta 3):

- savepoint em volta das escritas condicionais;
- travar todas as linhas antes de escrever;
- gravar a recusa numa transação nova.

Repetição (discovery, pergunta 12):

- sem repetição no servidor;
- executor repete deadlock e serialização; o resto vira 503;
- executor repete também espera de lock e falha de conexão.

## Decisão

Escolhida: **update condicional em `READ COMMITTED`**, porque:

- confere e grava num comando só, sem janela entre ler e escrever;
- não gera tentativas desperdiçadas na disputa do recreio;
- é a técnica que o Modern Treasury recomenda para "não deixar o saldo negativo" (*balance locking*).

### Débito e crédito

```sql
UPDATE ledger.accounts
SET balance = balance + $delta, version = version + 1
WHERE id = $id
  AND NOT ($direction <> normal_balance AND debit_blocked)
  AND (allow_negative OR balance + $delta >= 0)
RETURNING balance, version;
```

- `$direction` é a direção do lançamento; `$delta` é o valor com sinal calculado pelo lado normal da conta (ADR 0005).
- O `UPDATE` trava a linha; a transação concorrente espera o commit e, em `READ COMMITTED`, reavalia o `WHERE` com o saldo novo.
- 0 linhas é recusa. Uma leitura, só nesse caminho, distingue o motivo:
  - saldo insuficiente;
  - conta bloqueada;
  - conta inexistente.
- O bloqueio fica na mesma condição do saldo, nunca numa checagem anterior, e vale só para lançamento que diminui o saldo (direção oposta ao lado normal). Na carteira, isso é o débito; créditos continuam entrando.
- `RETURNING` fornece o saldo resultante e a versão gravados no lançamento.
- `CHECK (allow_negative OR balance >= 0)` na tabela como segunda barreira.
- O limite diário (spec 0003), outra linha disputada, usa a mesma técnica.

### Ordem e recusa dentro da transação

- Ordem global de travamento, igual em compra e estorno, para não haver deadlock entre linhas de tipos diferentes:
  1. linha do gasto do dia do aluno (limite diário);
  2. contas do ledger, em ordem de ID.
- O estorno abate o gasto do dia antes do lançamento inverso, na mesma ordem.
- As escritas condicionais (gasto do dia e ledger) rodam dentro de um savepoint. Com as contas em ordem de ID, a receita pode ser creditada antes de a carteira recusar; na recusa, a transação volta ao savepoint e o canteen grava a tentativa recusada, a auditoria e o aviso na mesma transação.
- O pgx abre savepoint ao iniciar uma transação dentro de outra. Um savepoint por compra não causa problema; o Postgres só degrada com dezenas de subtransações abertas na mesma transação.

### Contas quentes

- Todas as contas, inclusive receita e entrada de Pix, com saldo materializado e atualizado de forma síncrona.
- Estimativa no discovery: a linha da receita fica ocupada ~0,3% do tempo no volume real e ~30% no pior caso (100 compras/s numa só cantina); satura perto de 300 compras/s numa conta.
- Medição na Fase 6 com cenário concentrado. Se a espera passar da meta, as saídas, em ordem:
  - conta quente sem saldo materializado, revendo a ADR 0005 para essas contas;
  - subcontas;
  - saldo da conta quente aplicado de forma assíncrona, como no Modern Treasury.

### Repetição e limites de espera

- O executor de transação repete a transação inteira só em `40001` (serialização) e `40P01` (deadlock): até 3 tentativas, backoff curto com *full jitter*. A transação só toca o banco; eventos saem pelo outbox.
- Espera de lock acima do limite (`55P03`), timeout de comando e falha de conexão viram 503 com `Retry-After`; o caixa repete com a mesma chave de idempotência, também com *full jitter*.
- Nunca são repetidos:
  - recusa por saldo, limite, regra ou bloqueio;
  - livro diferente;
  - conflito de idempotência.
- Limites configurados na role da aplicação, nunca no `postgresql.conf` (as migrations precisam de locks longos). Pontos de partida, a confirmar no teste da RNF-CARGA-03 e na Fase 6:
  - `lock_timeout`: 500 ms, bem acima da fila máxima estimada da RNF-CARGA-03 (~20 × 3 ms);
  - `statement_timeout`: 2 s, acima do `lock_timeout`, senão o `lock_timeout` não tem efeito;
  - `idle_in_transaction_session_timeout`: 5 s.
- Cada repetição gera métrica e log com o código do erro (RNF-OBS-02).

### Verificação

- Teste com 20 goroutines e barreira de largada debitando a mesma carteira: saldo nunca negativo, aprovadas = ⌊saldo inicial / valor⌋, nenhum erro técnico.
- Teste com compra e estorno simultâneos do mesmo aluno, passando pelo gasto do dia e pelo ledger, para deadlock.
- Teste de recusa por saldo em que a receita é atualizada antes da carteira: a recusa é gravada, sem erro técnico e sem lançamento.
- Verificação de invariantes ao fim de todo teste de integração.

## Consequências

- Positivas:
  - A regra de saldo e de bloqueio está num `WHERE` só, visível e testável.
  - Disputa vira espera curta, não erro nem repetição.
  - Recusa por saldo sai como "0 linhas", tratada como resultado de negócio.
  - Saldo resultante e versão vêm do mesmo comando, como a ADR 0005 exige.
  - Nenhuma peça nova além do Postgres.
  - Trocar de técnica depois é mudar o adaptador, sem migração.
- Negativas:
  - Toda regra que precisar de garantia sob concorrência tem que caber numa condição sobre a própria linha; regra que dependa de várias linhas lidas juntas exige rever a técnica.
  - Uma checagem prévia em Go, acrescentada por engano, reabre a janela de *lost update* sem erro visível; a proteção é o teste de concorrência em todo `make check`.
  - A recusa precisa de uma leitura extra para dizer o motivo.
  - A receita é uma linha compartilhada por todas as compras da cantina; o limite está estimado, não medido, até a Fase 6.
  - Os valores de limite de espera são pontos de partida e podem precisar de ajuste.
  - Um savepoint por compra e uma ordem global de travamento a respeitar em todo fluxo novo que toque o gasto do dia e o ledger.

## Prós e contras das opções

### Update condicional em `READ COMMITTED`

- Prós:
  - Um comando por conta, sem janela entre ler e gravar.
  - Nenhuma tentativa desperdiçada: a transação concorrente espera e reavalia.
  - Recusa é resultado, não exceção.
  - Técnica recomendada pelo Modern Treasury; equivalente às flags de saldo do TigerBeetle.
- Contras:
  - 0 linhas não diz o motivo da recusa.
  - Deadlock possível sem ordem fixa das contas.
- Limitações: serve a regras que cabem numa condição sobre a própria linha.

### `SELECT ... FOR UPDATE` em ordem de ID e checagem em Go

- Prós:
  - Explícito: trava, confere, grava.
  - Acomoda regras que dependem de várias linhas lidas juntas.
  - Usado pelo pgledger (`pgledger_create_transfers` ordena as contas e trava uma a uma).
- Contras:
  - Dois comandos por conta.
  - A garantia depende de nunca esquecer o `FOR UPDATE`.
- Limitações: a mesma disputa do update condicional.

### Versão otimista

- Prós:
  - Não segura lock entre a leitura e a escrita.
  - Serve a "o cliente viu o saldo X e confirma sobre X" (`lock_version` do Modern Treasury).
- Contras:
  - Na conta disputada, repetições em série; o Modern Treasury mostra seis chamadas contra duas da condição de saldo, e transações que nunca conseguem gravar em conta quente.
  - Quem chama precisa repetir.
- Limitações: protege contra mudança, não contra saldo negativo.

### `SERIALIZABLE` com repetição

- Prós:
  - O código fica simples, como se não houvesse concorrência.
  - Protege também regras entre várias tabelas.
- Contras:
  - Toda transação precisa de laço de repetição.
  - Na linha disputada, a taxa de aborto cresce e a latência fica imprevisível, exatamente o cenário da RNF-CARGA-03.
- Limitações: o custo de detecção de conflito existe em toda transação, não só nas disputadas.
