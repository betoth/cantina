# 0005. Modelo contábil do ledger

- Status: aceita
- Data: 2026-10-06

## Contexto e problema

O ledger registra a quem pertence cada centavo da cantina e é permanente e append-only (RNF-RET-01): o que for gravado nunca é reescrito. Por isso a forma de representar contas, transações, lançamentos e saldos precisa estar certa antes do primeiro dado, porque mudá-la depois exige migrar um histórico que, por desenho, ninguém altera. O modelo também define o contrato da porta do ledger, usado por compra, estorno, recarga e extrato.

## Fatores de decisão

- Invariantes verificáveis pelo banco, não só pelo código: toda transação fecha (débitos = créditos), e saldo materializado = soma dos lançamentos.
- Leitura direta: saldo e extrato compreensíveis sem conhecer convenção interna; saldo negativo indica erro.
- Extensível sem refazer o modelo: transação com mais de duas contas (ex.: taxa da plataforma), novos tipos de conta.
- Histórico reconstruível: saldo de qualquer conta em qualquer momento, sem recalcular tudo.
- Precisão: valores exatos, sem arredondamento em nenhuma camada (banco, Go, JSON).
- Mercado: modelo reconhecido em contabilidade e em ledgers de fintech, fácil de defender e de comparar com referências (Modern Treasury, TigerBeetle).

## Opções consideradas

Como representar o lançamento e ler o saldo da conta:

- Valor com sinal e convenção única
- Direção e valor positivo no lançamento, lado normal na conta
- Débitos e créditos acumulados separados na conta

## Decisão

Escolhida: **direção e valor positivo no lançamento, lado normal na conta**, porque é a única das três em que todo saldo saudável é positivo e todo negativo é erro, com o vocabulário da contabilidade e do Modern Treasury. O custo (`CASE` nas somas) fica em poucas consultas do adaptador.

### Estrutura

- **Livro** (`Book`): conjunto fechado de contas. Transação só entre contas do mesmo livro; o ledger recusa o contrário. Um livro por escola no MVP; o ledger guarda só o ID, e quem sabe que o livro é de uma escola é o canteen. Mesmo papel do `ledger` do TigerBeetle e dos *shelves* do Square Books.
- **Conta**: livro, lado normal, se pode ficar negativa, saldo e versão. Não conhece o dono.
- **Transação**: o fato (ID gerado por quem chama, livro, código do motivo, transação que ela reverte). Dois ou mais lançamentos, com total de débitos = total de créditos.
- **Lançamento**: conta, direção, valor, saldo resultante e versão da conta.
- Transação com N lançamentos, não transferência de-para com duas pernas (como no Fluxo-De-Caixa): duas pernas não comportam taxa da plataforma nem pagamento com duas fontes, e o domínio já diz "dois ou mais".

### Lado normal

Definido na abertura da conta por quem chama, conforme o que a conta é para a cantina (dona do livro): o que ela tem cresce no débito; o que deve e o que ganha crescem no crédito.

| Conta | Natureza | Lado normal |
|---|---|---|
| entrada de Pix | ativo: dinheiro no banco da cantina | débito |
| carteira do aluno | passivo: crédito pré-pago devido ao aluno | crédito |
| receita da cantina | receita | crédito |

Efeito no saldo: direção do lançamento igual ao lado normal soma; diferente subtrai.

| Passo | Lançamentos | Entrada de Pix | Carteira | Receita |
|---|---|---|---|---|
| recarga de R$ 120 | D entrada de Pix / C carteira | 120 | 120 | 0 |
| compra de R$ 45 | D carteira / C receita | 120 | 75 | 45 |
| estorno da compra | D receita / C carteira | 120 | 120 | 0 |

Invariantes: cada transação tem débitos = créditos; por livro, total de débitos = total de créditos; saldo da conta = soma dos seus lançamentos pelo lado normal; nenhuma conta fica negativa, salvo as marcadas como podendo.

### Histórico

- Cada lançamento grava o saldo resultante da conta e a versão da conta (que sobe 1 a cada lançamento), com versão única por conta. O extrato mostra o saldo após cada movimento sem recalcular, o saldo em qualquer momento é uma leitura, e versão pulada ou repetida denuncia erro.
- Instantes em `timestamptz` (UTC). O ledger não tem noção de dia; dia, fuso e período são do canteen (ex.: limite diário no fuso da escola). A ordem dos lançamentos de uma conta é dada pela versão, não pelo relógio.

### Valores

- Inteiro (`bigint` no Postgres, `int64` em Go) na menor unidade da moeda: centavos. R$ 12,34 = `1234`. Entrada e exibição convertem dividindo ou multiplicando por 100; dentro do ledger só há soma e subtração de inteiros.
- Moeda única (real), expoente fixo 2: sem coluna de moeda nem de expoente.
- Sem `float` (0,1 + 0,2 = 0,30000000000000004) e sem `decimal`: Go não tem tipo decimal nativo, e cada conversão é uma chance de arredondar. Inteiro é exato no banco, em Go e em JSON.
- Mesmo padrão do Modern Treasury (`amount` na menor unidade, com `currency_exponent`), do Stripe (centavos) e do TigerBeetle (inteiro com escala fixa por ledger).
- Cálculo com percentual (ex.: taxa da plataforma, v2) gera frações de centavo: a regra de arredondamento é aplicada no cálculo, antes de gravar. O ledger só recebe centavos inteiros.

## Consequências

- Positivas: saldos legíveis sem conhecer convenção; "nenhuma conta negativa" serve de checagem única na reconciliação e em alertas; transações com mais de duas contas sem mudar o modelo; extrato e saldo histórico direto dos lançamentos; dinheiro exato em todas as camadas; modelo comparável campo a campo com o Modern Treasury.
- Negativas: `CASE` pelo lado normal nas somas de saldo e de verificação; lado normal é uma propriedade a mais que quem abre a conta precisa acertar, mitigado por tabela fixa no canteen coberta por teste e pelo `CHECK` de saldo, que recusa o primeiro lançamento de uma conta aberta com o lado errado; saldo resultante e versão em cada lançamento são redundância a manter coerente (coberta pela reconciliação).

## Prós e contras das opções

### Valor com sinal e convenção única

Lançamento com valor positivo (crédito) ou negativo (débito); saldo de toda conta = créditos − débitos. Usado por pgledger e Formance (a conta `@world`, origem de todo dinheiro, fica negativa).

- Prós: soma direta em SQL, sem `CASE`; uma regra só para todas as contas; soma zero vira `sum(amount) = 0`.
- Contras: contas de ativo ficam negativas no uso normal (a entrada de Pix com R$ 100 no banco aparece −100); o sinal sozinho não diz se há erro, então "pode ficar negativa" vira flag de toda conta de sistema; leitura depende de conhecer a convenção.

### Direção e valor positivo no lançamento, lado normal na conta

Lançamento com `direction` (`debit` ou `credit`) e `amount > 0`; conta com `normal_balance`. Direção igual ao lado normal soma; oposta subtrai. Modelo da contabilidade e do Modern Treasury (`normal_balance` e `lock_version` na conta; `direction`, `amount` e `resulting_ledger_account_balances` no lançamento).

- Prós: toda conta saudável fica positiva, e negativo é sempre anomalia (uma checagem só na reconciliação); cada linha se lê sozinha; mesmo vocabulário da contabilidade e da referência de mercado mais usada.
- Contras: `CASE` em toda soma (saldo e soma zero); uma propriedade a mais na abertura de conta, que quem chama precisa acertar.

### Débitos e créditos acumulados separados na conta

Conta guarda `debits_posted` e `credits_posted`; o saldo é calculado por quem lê. Modelo do TigerBeetle, que usa flags (`debits_must_not_exceed_credits`) no lugar do lado normal.

- Prós: nenhuma convenção gravada; qualquer leitura de saldo é possível; acumulados só crescem.
- Contras: a regra "não pode ficar negativa" vira comparação entre duas colunas, e o saldo não existe como valor pronto para o update condicional; o lado continua precisando estar em algum lugar (flag ou código de quem lê); mais distante do vocabulário de quem lê o extrato.
