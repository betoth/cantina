# Discovery: ledger

- Data: 2026-10-05
- Motivação: decisões da spec 0001 (ledger) que mudam o modelo e são caras de reverter. O ledger é o maior risco técnico do projeto (Fase 1).
- Referências pesquisadas: TigerBeetle, Modern Treasury, Formance, Square Books, Stripe, Uber LedgerStore, pgledger, SQL Server Ledger, dois projetos de estudo (Fluxo-De-Caixa, simple-ledger) e artigos sobre Postgres. Lista completa em [Fontes](#fontes).

Cada pergunta traz como o mercado resolve, as opções e a recomendação para a cantina. Recomendação não é decisão: as decisões saem nas ADRs, no `domain.md` e na spec (ver [Saídas](#saídas)).

## Resumo das recomendações

| # | Pergunta | Recomendação |
|---|---|---|
| 1 | Partição | livro (`Book`) como ID opaco; um por escola no MVP |
| 2 | Concorrência no saldo | update condicional (balance locking), `READ COMMITTED` |
| 3 | Contas quentes | sem tratamento especial no MVP; atualizar saldos no fim da transação; medir na Fase 6 |
| 4 | Duas fases | não no MVP |
| 5 | Saldos | um saldo só (lançado), pelo lado normal da conta |
| 6 | Fronteira de dados | ledger guarda só o contábil; motivo por código; detalhes no canteen |
| 7 | Idempotência | ID da transação gerado por quem chama é a chave; conteúdo diferente é erro |
| 8 | Imutabilidade e verificação | permissões por coluna, soma zero no banco, saldo e versão em cada lançamento; sem hash encadeado |
| 9 | Construir ou usar pronto | construir sobre o Postgres, copiando os padrões abaixo |
| 10 | Schema | `books`, `accounts`, `transactions`, `entries`; direção e valor positivo em `bigint` (centavos); UUIDv7 |
| 11 | Idempotência no banco | chave primária + `ON CONFLICT DO NOTHING` + comparação dos campos |
| 12 | Retentativa | transação inteira, no chamador, poucas tentativas com jitter; `lock_timeout` |
| 13 | Erros | erros de domínio exportados no pacote do ledger; canteen traduz; HTTP em RFC 9457 |
| 14 | Transação compartilhada | executor de transação na aplicação, portas ligadas à transação explicitamente |
| 15 | Testes | concorrência com testcontainers, testes de propriedade e verificação de invariantes |

## Modelo

### 1. Partição: livro

**Mercado.** TigerBeetle tem o campo `ledger` em conta e transferência: só contas do mesmo ledger transacionam diretamente. Normalmente um ledger por moeda; em sistemas multi-tenant, um por cliente ou por cliente e moeda. Formance permite vários ledgers independentes por instalação. Square Books isola tenants em *shelves*. É padrão: partição contábil fechada, sem dono de negócio.

**Para nós.**

- Livro = conjunto fechado de contas. Transação só entre contas do mesmo livro; o ledger recusa o contrário.
- O ledger guarda só o `book_id`; quem sabe que o livro X é a escola Y é o canteen.
- Invariante 2 (soma dos saldos é zero) passa a valer por livro, o que permite reconciliar escola a escola.
- Granularidade segue quem guarda o dinheiro. MVP: uma cantina por escola, um livro por escola. Cantinas de donos diferentes na mesma escola exigiriam livro por cantina (o aluno teria uma carteira em cada) ou recebimento pela plataforma com repasse (UC-SIS-04, v2). Questão de domínio.

**Recomendação:** livro por escola, `book_id` em conta e transação, checado na escrita. Registrar no `domain.md` a regra "o livro segue quem guarda o dinheiro" e a questão das cantinas de donos diferentes.

### 2. Concorrência no saldo

**Mercado.** Três técnicas:

| Técnica | Como | Quem usa |
|---|---|---|
| Lock pessimista | `SELECT ... FOR UPDATE` nas contas em ordem fixa, confere, grava | pgledger (ordena as contas e trava uma a uma) |
| Versão otimista | cliente manda a versão lida (`lock_version`); grava só se não mudou; senão, falha e retenta | Modern Treasury (opcional, para quem precisa de "nada mudou desde que li") |
| Condição de saldo (*balance locking*) | grava só se o saldo resultante ficar na faixa exigida (ex.: `>= 0`) | Modern Treasury (recomendado), TigerBeetle (flags `debits_must_not_exceed_credits`) |

Modern Treasury migrou de versão para condição de saldo por causa das contas quentes: com versão, duas compras de R$ 250 e R$ 750 contra R$ 1.000 custam seis chamadas (lê, tenta, falha, relê); com condição de saldo, duas. Testes publicados no Postgres mostram o lock pessimista cerca de 2× mais rápido que o otimista `SERIALIZABLE` numa linha disputada, sem tentativas desperdiçadas.

**Anti-padrão observado (Fluxo-De-Caixa).** Lock distribuído no Redis (Redlock) nas duas contas, saldo lido, somado em memória e gravado pelo ORM. Três falhas:

- O lock é liberado ao sair do método que o adquire (`await using`), antes da gravação: não protege nada, e o saldo sofre *lost update*.
- Lock externo com expiração e sem *fencing token* não garante exclusão: uma pausa de GC ou de rede maior que a expiração deixa dois processos gravando. A consistência do saldo tem que vir do próprio banco, na mesma transação.
- Contas travadas na ordem origem → destino: A→B e B→A simultâneos esperam um pelo outro até o timeout.

Confirma a escolha de lock no Postgres (linha do `UPDATE`) e a ordem fixa por ID.

**Para nós.** Update condicional é a condição de saldo dentro do Postgres:

```sql
UPDATE ledger.accounts
SET balance = balance + $delta, version = version + 1
WHERE id = $id AND (allow_negative OR balance + $delta >= 0)
RETURNING balance, version;
-- 0 linhas: saldo insuficiente (ou conta inexistente; distinguir depois)
```

- O `UPDATE` pega o lock da linha sozinho; não há janela entre ler e escrever.
- `READ COMMITTED` (padrão do Postgres): o segundo `UPDATE` espera o primeiro e reavalia o `WHERE` com o valor já confirmado. Sem erro de serialização.
- Contas atualizadas sempre na mesma ordem (por ID) para não haver deadlock.
- `CHECK (allow_negative OR balance >= 0)` como segunda barreira.

**Recomendação:** update condicional em `READ COMMITTED`, ordem por ID, `CHECK` no banco. Versão otimista fica disponível de graça (a coluna `version` existe), mas não é exposta a quem chama no MVP. Vira ADR.

### 3. Contas quentes

**Mercado.** Conta que participa de muitas transações (receita, entrada de Pix) vira fila na linha do saldo. Soluções, da mais simples à mais pesada: segurar o lock o mínimo possível (gravar lançamentos primeiro, atualizar saldos por último, em ordem de ID); não materializar o saldo da conta quente e calculá-lo sob demanda ou de forma assíncrona (Modern Treasury atualiza versões em background quando o cliente não pede `lock_version`); dividir a conta em subcontas (*sharding*); enfileirar e gravar em lote (TigerBeetle). pgledger mediu cerca de 7.500 transferências/s com 10 contas disputadas num notebook: contenção em Postgres só vira problema bem acima do nosso volume.

**Para nós.** Receita da cantina e entrada de Pix são quentes. Volume real: menos de 1 compra/s por cantina. RNF-CARGA-01 (100 compras/s) é a soma de cerca de 100 escolas, então cerca de 1/s por conta de receita. Risco aparece só se o cenário de carga concentrar tudo numa cantina.

**Recomendação:** materializar o saldo de todas as contas; dentro da transação, gravar lançamentos primeiro e atualizar saldos por último, em ordem de ID. Medir na Fase 6. Se a receita virar gargalo, a primeira saída é não materializar o saldo de contas que só recebem crédito na compra.

### 4. Duas fases (pendente → lançada ou anulada)

**Mercado.** TigerBeetle: transferência pendente reserva o valor (`debits_pending`), depois é lançada (total ou parcial), anulada ou expira. Modern Treasury: transação nasce `pending` e passa a `posted` quando o dinheiro se move no banco. Serve a reservas: autorização de cartão, escrow, pagamento que depende de sistema externo.

**Para nós.**

- Compra: autorização e captura no mesmo instante, na mesma transação. Não precisa de reserva.
- Recarga: o crédito só nasce quando o Pix é confirmado. A cobrança pendente vive no payments e no pedido de recarga do canteen, não no ledger.
- Primeiro caso real seria encomenda antecipada (UC-RESP-12, v2): reservar o valor até a retirada.

**Recomendação:** não no MVP. Adicionar depois exige status na transação e saldo pendente na conta: migration e mudança de código, mas sem refazer o modelo. Registrar como fora de escopo da spec.

### 5. Saldos

**Mercado.** Modern Treasury expõe três: lançado (*posted*), pendente (lançado + pendente) e disponível (entradas lançadas − saídas lançadas e pendentes). TigerBeetle guarda os quatro acumulados (débitos e créditos, pendentes e lançados) e o saldo sai deles.

**Para nós.** Sem duas fases, os três coincidem.

**Lado normal.** Duas formas de ler o saldo:

| Forma | Como | Quem usa |
|---|---|---|
| Convenção única | saldo = créditos − débitos em toda conta; conta de origem do dinheiro fica negativa | Formance (`@world` negativa), pgledger |
| Lado normal | cada conta tem `normal_balance` (`debit` ou `credit`); lançamento na direção igual soma, na oposta subtrai | contabilidade, Modern Treasury (`normal_balance` na conta; `direction` e `amount` positivo no lançamento), simple-ledger |

TigerBeetle não escolhe: guarda `debits_posted` e `credits_posted` separados, e as flags `debits_must_not_exceed_credits` / `credits_must_not_exceed_debits` fazem o papel do lado normal.

Na cantina, o lado sai do que a conta é para a cantina (dona do livro): o que ela tem (ativo) cresce no débito; o que deve (passivo) e o que ganha (receita) crescem no crédito.

| Conta | Natureza | Lado normal |
|---|---|---|
| entrada de Pix | ativo: dinheiro no banco da cantina | débito |
| carteira do aluno | passivo: crédito pré-pago devido ao aluno | crédito |
| receita da cantina | receita | crédito |

Exemplo (recarga de R$ 120, compra de R$ 45, estorno da compra), saldos pelo lado normal:

| Passo | Lançamentos | Entrada de Pix | Carteira | Receita |
|---|---|---|---|---|
| recarga | D entrada de Pix 120 / C carteira 120 | 120 | 120 | 0 |
| compra | D carteira 45 / C receita 45 | 120 | 75 | 45 |
| estorno | D receita 45 / C carteira 45 | 120 | 120 | 0 |

Conferência a qualquer momento: entrada de Pix = carteiras + receita. Na convenção única a entrada de Pix apareceria −120: mesma informação, leitura contraintuitiva (a dúvida apareceu na própria discussão: "entrou Pix, não deveria ser crédito?").

Ganhos do lado normal: toda conta saudável fica positiva; **negativo é sempre anomalia**, então uma checagem só ("nenhuma conta negativa") serve à reconciliação e a alerta, e `allow_negative` vira exceção rara em vez de flag de toda conta de sistema; modelo igual ao do Modern Treasury, fácil de defender. Custo: uma coluna na conta e um `CASE` nas somas.

**Recomendação:** um saldo só, lido pelo lado normal da conta. `normal_balance` definido na abertura da conta por quem chama (o canteen sabe que é carteira; o ledger só guarda a propriedade contábil, como já faz com "pode ficar negativa"). Lançamento com direção e valor positivo. Invariante 2 passa de "soma dos saldos é zero" para "por livro, total de débitos = total de créditos". Muda o `domain.md` ("a conta de entrada de Pix fica negativa por definição").

### 6. Fronteira de dados

**Mercado.** TigerBeetle guarda só o contábil e oferece campos opacos para ligar ao sistema de origem: `user_data_128` (quem/o quê), `user_data_64` (segundo carimbo de tempo), `user_data_32` (onde), e `code` (por quê: tipo de conta ou motivo da transferência). Formance e pgledger aceitam metadados livres (JSON). Stripe e Square tratam o ledger como registro de fatos financeiros; o detalhe de negócio fica nos sistemas de origem.

**Para nós.** O domínio já decidiu que o ledger não conhece donos.

| No ledger | No canteen |
|---|---|
| livro, contas, se a conta pode ficar negativa | de quem é cada conta (aluno, cantina) |
| transação: ID, código do motivo (`purchase`, `refund`, `top_up`), transação que ela reverte, instante | compra: itens, preços, categorias, versão das regras, operador, terminal |
| lançamentos: conta, valor, saldo resultante, versão da conta | auditoria, eventos, avisos |

**Recomendação:** sem metadados livres (JSON) no ledger: tudo o que é de negócio fica no canteen, que guarda o ID da transação do ledger. O código do motivo fica no ledger porque ajuda a reconciliar e a ler o extrato sem consultar o canteen.

### 7. Idempotência

**Mercado.**

- TigerBeetle: o ID da transferência, gerado pelo cliente, **é** a chave de idempotência. Repetir devolve `exists`; repetir com campos diferentes devolve `exists_with_different_*` (erro). Recomenda IDs ordenáveis por tempo.
- Stripe: chave no header `Idempotency-Key`; guarda status e corpo da primeira resposta (até erro 500) e devolve igual; parâmetros diferentes com a mesma chave é erro; chaves podem ser removidas após 24 h.
- Formance: chave de idempotência por transação.
- Anti-padrão (simple-ledger): chave = hash do conteúdo da requisição, em cache na memória do processo com TTL de 15 min. Falha com vários processos ou após restart, e recusa operações repetidas legítimas: o mesmo aluno comprar o mesmo item duas vezes em 15 min seria tratado como duplicata. A chave tem que ser gerada por quem chama e persistida com unicidade, na mesma transação.

**Para nós.** Há duas camadas, que não se confundem:

- **Requisição** (spec 0004): o caixa reenvia a compra; o canteen devolve o resultado gravado, inclusive recusa. Precisa guardar a resposta.
- **Ledger** (esta spec, invariante 5): quem chama gera o ID da transação. Repetir com o mesmo conteúdo devolve a transação existente; com conteúdo diferente, erro. Protege também o consumidor de recarga, que recebe eventos pelo menos uma vez (Fase 2).

**Recomendação:** padrão TigerBeetle no ledger. Escopo global (UUID não colide entre livros). Sem expiração: o ledger é permanente (RNF-RET-01).

### 8. Imutabilidade e verificação

**Mercado.**

- Modern Treasury: versão por conta (`lock_version`) que sobe a cada lançamento; cada lançamento aponta a versão e o saldo resultante, o que permite reconstruir o saldo em qualquer momento. Distingue `effective_at` (quando o fato econômico aconteceu) de `created_at` (quando foi gravado).
- pgledger: cada lançamento guarda saldo anterior, saldo atual e versão da conta.
- Stripe: contas de compensação que devem zerar; o que não zera aponta erro. Prioriza detectar e delimitar o problema a impedir todo erro.
- Uber: verificação por checksum em janelas de tempo pegou bugs que lançamentos isolados não mostravam.
- Hash encadeado (cada linha guarda o hash da anterior): torna edição evidente, mas quem tem escrita total reescreve a cadeia inteira. Mais comum em trilha de auditoria do que em ledger operacional.
- SQL Server Ledger (2022+): tabela `APPEND_ONLY` (o banco recusa `UPDATE` e `DELETE`), hash encadeado por transação do banco e *digest* exportado periodicamente para armazenamento imutável fora do banco. O digest externo fecha a brecha do hash encadeado: quem reescreve a cadeia não consegue reescrever o digest já publicado. Irreversível depois de ligado. Sem equivalente nativo no Postgres; a versão manual seria hash encadeado, job de verificação e digest publicado fora. Custo: o encadeamento global serializa todas as escritas do ledger (por conta reduz a disputa, mas complica a verificação). Usado pelo Fluxo-De-Caixa.
- Saldo resultante gravado em cada lançamento também aparece no Fluxo-De-Caixa (`BalanceAfterTransaction`).
- Anti-padrões (simple-ledger): o repositório faz *upsert* da transação e apaga e reinsere os lançamentos (trigger só contra `DELETE` de transação e conta, nada contra `UPDATE`); e grava transação e lançamentos sem transação do banco, então uma queda no meio deixa transação desbalanceada. Imutabilidade declarada não basta: tem que valer para todas as tabelas e operações, e a escrita tem que ser atômica.

Separar duas proteções: **impedir** (permissões e triggers barram a aplicação, contra bug e invasão pela aplicação) e **detectar** (hash e digest externo revelam alteração feita por quem passou por cima da aplicação: DBA, superuser, acesso ao disco).

**Para nós.**

- Permissões: a role da aplicação tem só `SELECT` e `INSERT` em transações e lançamentos; em contas, `UPDATE` apenas nas colunas `balance` e `version` (permissão por coluna do Postgres).
- Soma zero no banco: *constraint trigger* adiado (`DEFERRABLE INITIALLY DEFERRED`) confere no commit que os lançamentos de cada transação somam zero. A invariante 1 deixa de depender só do Go.
- Saldo resultante e versão em cada lançamento, com `UNIQUE (account_id, account_version)`: histórico do saldo de graça (o extrato mostra o saldo após cada movimento) e verificação de que nenhuma versão foi pulada.
- Reconciliação: consulta "saldo materializado = soma dos lançamentos" por conta e "soma dos saldos = 0" por livro. Nos testes desde a Fase 1; agendada na Fase 2 (UC-SIS-02).

**Recomendação:** os quatro itens acima. Sem hash encadeado no MVP: as permissões, o trigger e a reconciliação cobrem o risco de bug, que é o risco real aqui; hash protege contra adulteração por quem tem acesso ao banco, ameaça fora do cenário de uma cantina, e o encadeamento global brigaria com o pico do recreio (RNF-CARGA). Candidato a v2 ("trilha com evidência de adulteração"), sempre com digest externo, sem o qual não protege. `effective_at` separado só quando houver lançamento retroativo (não há).

### 9. Construir ou usar pronto

**Mercado.** TigerBeetle: banco dedicado a contabilidade, muito rápido, com contas, transferências, duas fases e idempotência prontos. Formance: serviço de ledger sobre Postgres, com API e linguagem própria (Numscript). Modern Treasury: produto pago.

**Para nós.** Os dois prontos rodam fora da transação do Postgres do canteen. A compra precisa gravar débito, compra, auditoria e outbox atomicamente (RNF-AUD-06, ADR 0004). Com ledger externo, isso vira coordenação entre sistemas (duas fases ou saga), mais complexo que o problema. O volume não pede banco dedicado. E o projeto existe para dominar essas decisões.

**Recomendação:** construir o ledger como módulo do canteen sobre o Postgres, copiando: livro, código do motivo, flags de saldo e ID do cliente como chave (TigerBeetle); condição de saldo e versão por conta (Modern Treasury); saldo resultante e versão em cada lançamento, contas em ordem fixa (pgledger). Registrar na ADR "Ledger como módulo isolado no canteen", já prevista na Fase 1.

## Técnico

### 10. Schema

**Mercado.** pgledger: `accounts` (saldo `numeric`, `version`, `allow_negative_balance`, `allow_positive_balance`, moeda), `transfers` (de, para, valor `> 0`) e `entries` (valor com sinal, saldo anterior e atual, versão). Modern Treasury: contas, transações com lançamentos, versão de conta em tabela separada. TigerBeetle: inteiros de 128 bits e *asset scale* fixo por ledger (avisa que não dá para mudar depois).

**Esboço para nós** (detalhes ficam na spec):

```sql
ledger.books        (id uuid PK, created_at timestamptz)
ledger.accounts     (id uuid PK, book_id FK, normal_balance direction,
                     allow_negative bool, balance bigint, version bigint, created_at,
                     CHECK (allow_negative OR balance >= 0))   -- balance pelo lado normal
ledger.transactions (id uuid PK, book_id FK, code text,
                     reverses_id uuid UNIQUE NULL FK transactions, created_at)
ledger.entries      (id bigint PK, transaction_id FK, account_id FK,
                     direction direction,                      -- enum debit | credit
                     amount bigint CHECK (amount > 0),
                     balance_after bigint, account_version bigint, created_at,
                     UNIQUE (account_id, account_version))
```

- Lançamentos (não transferência de-para): uma transação pode ter mais de dois (o domínio diz "dois ou mais").
- Direção e valor positivo, como o Modern Treasury e a contabilidade: cada linha se lê sozinha. Custa um `CASE` nas somas (soma zero: débitos = créditos por transação; saldo: direção igual ao lado normal soma, oposta subtrai). Alternativa descartada: valor com sinal (pgledger), soma direta mas leitura dependente de convenção (pergunta 5).
- Correspondência com o Modern Treasury: `normal_balance` e `lock_version` na conta; `direction`, `amount` na menor unidade da moeda, `resulting_ledger_account_balances` e `ledger_account_lock_version` no lançamento.
- Dinheiro em inteiro de centavos, nunca ponto flutuante nem `decimal` de ORM (simple-ledger acerta; Fluxo-De-Caixa usa `decimal`).
- `reverses_id UNIQUE`: invariante 6 garantida pelo banco.
- IDs UUIDv7: gerados por quem chama, ordenáveis por tempo (recomendação do TigerBeetle), índice B-tree sem fragmentação.
- FK só dentro do schema `ledger`; nenhuma para outros contextos.
- Livro da transação igual ao livro das contas: conferido na escrita (Go e trigger).

### 11. Idempotência no banco

**Mercado.** Stripe/Brandur: tabela de chaves com `UNIQUE (usuário, chave)`, parâmetros da requisição e resposta; parâmetros diferentes retornam 409; requisições simultâneas com a mesma chave: uma trava, as outras recebem conflito. TigerBeetle: chave é a chave primária; conteúdo diferente é erro.

**Para nós** (camada do ledger):

1. `INSERT INTO ledger.transactions ... ON CONFLICT (id) DO NOTHING RETURNING id`.
2. Inseriu: segue com lançamentos e saldos.
3. Não inseriu: lê a transação existente e compara livro, código, reversão e lançamentos. Igual: devolve a existente sem efeito. Diferente: `ErrIdempotencyConflict`.
4. Duas chamadas simultâneas com o mesmo ID: a segunda espera o commit da primeira no `INSERT` (índice único) e cai no passo 3.

Sem hash do pedido: o ledger já guarda todos os campos, a comparação é direta. Hash e resposta guardada são da camada da requisição (spec 0004).

### 12. Retentativa

**Mercado.** Postgres não retenta sozinho e manda retentar a **transação inteira**, incluindo a lógica que decidiu o SQL: sempre em `40001` (serialização), recomendado em `40P01` (deadlock), com cuidado em `23505`. Stripe retenta no cliente com backoff exponencial e jitter, sempre com a mesma chave; erro de validação não é retentado.

**Para nós.**

| Camada | Retenta o quê | Como |
|---|---|---|
| Ledger | nada | roda dentro da transação de quem chama; não pode retentar sozinho |
| Canteen (executor de transação) | `40P01`, `40001`, `55P03` (`lock_timeout`) | transação inteira, até 3 tentativas, backoff curto com jitter |
| Cliente (caixa) | falha de rede, timeout, 503 | mesma chave de idempotência (spec 0004) |

- `READ COMMITTED` com ordem fixa praticamente elimina `40001` e `40P01`; a retentativa é rede de segurança.
- `lock_timeout` (ex.: 2 s) e `statement_timeout` limitam a espera numa conta disputada; estourou, vira erro transitório (503 com `Retry-After`) em vez de pendurar o caixa.
- Saldo insuficiente, livro diferente e conflito de idempotência são resultado, não falha: nunca retentados.

### 13. Erros

**Mercado.** TigerBeetle devolve um código por resultado (`exceeds_credits`, `exists_with_different_amount`, `accounts_must_have_the_same_ledger`, `pending_transfer_already_posted`). RFC 9457 padroniza erro HTTP em `application/problem+json`: `type` (URI estável), `title`, `status`, `detail`, `instance`, mais campos de extensão; e proíbe expor detalhes internos (stack, SQL).

**Para nós.**

- Erros de domínio exportados no pacote do ledger (convenção do projeto): `ErrInsufficientFunds`, `ErrAccountNotFound`, `ErrBookMismatch`, `ErrUnbalanced`, `ErrIdempotencyConflict`, `ErrAlreadyReversed`. `ErrUnbalanced` é bug de quem chama, não algo que o usuário veja.
- O ledger não fala HTTP. O canteen traduz para o seu domínio: `ErrInsufficientFunds` vira recusa da compra por saldo (gravada, com aviso); `ErrBookMismatch` vira erro interno.
- HTTP no formato RFC 9457, com campo `code` estável (ex.: `insufficient_funds`) para o cliente decidir sem ler texto. Se a compra recusada é erro HTTP ou resultado com status `recusada` é decisão da spec 0002 (ela é gravada e devolvida na retentativa, o que puxa para resultado).
- Erro técnico: 500 genérico com o ID do trace; detalhe só no log, sem dado pessoal (RNF-PRIV-05, RNF-OBS-04).

### 14. Transação compartilhada

**Mercado (Go).** Três padrões:

| Padrão | Como | Contra |
|---|---|---|
| Transação na assinatura | método da porta recebe `pgx.Tx` | infraestrutura vaza para a porta |
| Transação no `context` | adaptador põe e tira a transação do `ctx` | implícito: esquecer de propagar o `ctx` grava fora da transação sem erro |
| Unidade de trabalho | `WithinTx(ctx, func(ctx, uow) error)`; `uow` expõe as portas já ligadas à transação | uma interface a mais |

sqlc gera `queries.WithTx(tx)`, que encaixa nos três.

**Recomendação:** unidade de trabalho. Explícito (o compilador mostra o que roda dentro da transação), domínio sem pgx, e o executor é o lugar natural da retentativa (pergunta 12). Vai para a ADR "Portas e adaptadores com DDD tático leve", já prevista.

### 15. Testes

**Mercado.** TigerBeetle testa por simulação determinística com falhas injetadas (VOPR). Projetos em Postgres usam teste de estresse concorrente e conferência de invariantes ao fim. Testes de propriedade geram sequências aleatórias de operações e conferem invariantes após cada uma.

**Para nós.**

- Domínio puro (tabela, unitário): transação balanceada, valores, mesmo livro, reversão.
- Integração (testcontainers): cada critério de aceite contra o Postgres real, incluindo as permissões (`UPDATE` em lançamento falha com a role da aplicação).
- Concorrência: N goroutines debitando a mesma carteira; ao fim, saldo nunca negativo, aprovadas = ⌊saldo inicial / valor⌋, nenhum erro técnico (RNF-CARGA-03, critério 2 da Fase 1).
- Invariantes: função de verificação (invariantes 1 a 4 e versões sem buraco) chamada ao fim de todo teste de integração; é a mesma consulta da reconciliação da Fase 2.
- Propriedade: sequências aleatórias de abrir conta, transacionar e reverter, conferindo as invariantes (lib `pgregory.net/rapid`). Opcional na primeira entrega; recomendado para o domínio puro.

## Saídas

| Saída | O quê |
|---|---|
| ADR nova | Concorrência no saldo: update condicional, `READ COMMITTED`, ordem por ID, retentativa no executor (perguntas 2, 3, 12) |
| ADR "Ledger como módulo isolado no canteen" (prevista) | construir vs pronto, livro, fronteira de dados, o que foi copiado de cada referência (perguntas 1, 6, 9) |
| ADR "Portas e adaptadores com DDD tático leve" (prevista) | unidade de trabalho para transação compartilhada (pergunta 14) |
| `domain.md` | glossário: Livro; regra "o livro segue quem guarda o dinheiro"; convenção de saldo; questão em aberto: cantinas de donos diferentes na mesma escola |
| Spec 0001 | schema, idempotência, erros, testes e fora de escopo (duas fases, hash encadeado, contas quentes) |
| Spec 0004 | idempotência da requisição: resposta guardada, hash do pedido, tratamento de chave simultânea |
| Spec 0002 | compra recusada como resultado ou erro HTTP; formato RFC 9457 |

## Fontes

- TigerBeetle: [Data modeling](https://docs.tigerbeetle.com/coding/data-modeling/), [Two-phase transfers](https://docs.tigerbeetle.com/coding/two-phase-transfers/), [Reliable transaction submission](https://docs.tigerbeetle.com/coding/reliable-transaction-submission/)
- Modern Treasury: [How to scale a ledger, part I](https://www.moderntreasury.com/journal/how-to-scale-a-ledger-part-i), [part IV](https://www.moderntreasury.com/journal/how-to-scale-a-ledger-part-iv), [Designing the Ledgers API with optimistic locking](https://www.moderntreasury.com/journal/designing-ledgers-with-optimistic-locking), [How to handle concurrent transactions](https://www.moderntreasury.com/journal/how-to-handle-concurrent-transactions), [Transaction status and balances](https://docs.moderntreasury.com/ledgers/docs/transaction-status-and-balances), [Verify prior ledger states](https://docs.moderntreasury.com/ledgers/docs/verify-prior-ledger-states), [Ledger account object](https://docs.moderntreasury.com/ledgers/reference/ledger-account-object), [Ledger entry object](https://docs.moderntreasury.com/ledgers/reference/ledger-entry-object)
- Formance: [Ledger](https://docs.formance.com/ledger)
- Square: [Books, an immutable double-entry accounting database service](https://developer.squareup.com/blog/books-an-immutable-double-entry-accounting-database-service/)
- Stripe: [Ledger](https://stripe.dev/blog/ledger-stripe-system-for-tracking-and-validating-money-movement), [Designing robust and predictable APIs with idempotency](https://stripe.com/blog/idempotency), [Idempotent requests](https://docs.stripe.com/api/idempotent_requests)
- Uber: [How LedgerStore supports trillions of indexes](https://www.uber.com/blog/how-ledgerstore-supports-trillions-of-indexes/)
- pgledger: [repositório](https://github.com/pgr0ss/pgledger), [A ledger in PostgreSQL is fast](https://www.pgrs.net/2025/05/16/pgledger-in-postgresql-is-fast/)
- Brandur: [Implementing Stripe-like idempotency keys in Postgres](https://brandur.org/idempotency-keys)
- Postgres: [Serialization failure handling](https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html)
- [RFC 9457: Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457.html)
- Transações em Go hexagonal: [Qonto](https://medium.com/qonto-way/transactions-in-go-hexagonal-architecture-f12c7a817a61), [Kaznacheev](https://www.kaznacheev.me/posts/en/clean-transactions-in-hexagon/)
- Contas quentes em Postgres: [payments-ledger](https://github.com/Arjun-B-J/payments-ledger)
- Hash encadeado: [Tamper-evident audit trails in PostgreSQL](https://appmaster.io/blog/tamper-evident-audit-trails-postgresql)
- SQL Server: [Ledger overview](https://learn.microsoft.com/sql/relational-databases/security/ledger/ledger-overview)
- Projetos de estudo (exemplos e anti-padrões, não referência): [Fluxo-De-Caixa](https://github.com/CristianoRC/Fluxo-De-Caixa) (.NET, SQL Server Ledger, Redlock), [simple-ledger-test-reference](https://github.com/gutogalego/simple-ledger-test-reference) (Node, SQLite, desafio de ledger)
