# Discovery: ledger

- Motivação: decisões da spec 0001 (ledger) que mudam o modelo e são caras de reverter. O ledger é o maior risco técnico do projeto (Fase 1).
- Histórico:
  - 2026-10-05: criação (perguntas 1 a 15).
  - 2026-10-06: análise de dois projetos de estudo e do SQL Server Ledger; lado normal (pergunta 5); adaptação ao template da skill `/discovery`; aprofundamento de concorrência (perguntas 2, 3 e 12).
- Referências pesquisadas: TigerBeetle, Modern Treasury, Formance, Square Books, Stripe, Uber LedgerStore, pgledger, SQL Server Ledger, documentação do Postgres, AWS, dois projetos de estudo (Fluxo-De-Caixa, simple-ledger). Lista completa em [Fontes](#fontes).

Cada pergunta traz como o mercado resolve, as opções comparadas e a recomendação para o projeto. Recomendação não é decisão: as decisões saem nas ADRs, no `domain.md` e nas specs (ver [Saídas](#saídas)).

## Resumo das recomendações

| # | Pergunta | Recomendação |
|---|---|---|
| 1 | Partição | livro (`Book`) como ID opaco; um por escola no MVP |
| 2 | Concorrência no saldo | update condicional (balance locking) em `READ COMMITTED`, com bloqueio para débito na mesma condição e contas em ordem de ID |
| 3 | Contas quentes | todas as contas com saldo materializado e síncrono; débito na ordem natural da compra; medir na Fase 6 com cenário concentrado |
| 4 | Duas fases | não no MVP |
| 5 | Saldos | um saldo só (lançado), pelo lado normal da conta |
| 6 | Fronteira de dados | ledger guarda só o contábil; motivo por código; detalhes no canteen |
| 7 | Idempotência | ID da transação gerado por quem chama é a chave; conteúdo diferente é erro |
| 8 | Imutabilidade e verificação | permissões por coluna, partida dobrada conferida no banco, saldo e versão em cada lançamento, reconciliação; sem hash encadeado |
| 9 | Construir ou usar pronto | construir sobre o Postgres, copiando os padrões abaixo |
| 10 | Schema | `books`, `accounts`, `transactions`, `entries`; direção e valor positivo em `bigint` (centavos); UUIDv7 |
| 11 | Idempotência no banco | chave primária + `ON CONFLICT DO NOTHING` + comparação dos campos |
| 12 | Retentativa e limites de espera | executor repete a transação inteira só em deadlock e serialização; espera longa vira 503 e o caixa repete com a mesma chave; limites na role da aplicação |
| 13 | Erros | erros de domínio exportados no pacote do ledger; canteen traduz; HTTP em RFC 9457 |
| 14 | Transação compartilhada | executor de transação na aplicação, portas ligadas à transação explicitamente |
| 15 | Testes | concorrência com testcontainers, testes de propriedade e verificação de invariantes |

## 1. Partição: livro

**Pergunta.** Como separar as contas de escolas diferentes dentro do ledger, de forma que nenhuma transação misture dinheiro de donos diferentes e que cada escola possa ser reconciliada sozinha? Errar aqui é caro: a partição está em toda conta e toda transação gravadas, e o histórico é permanente.

**Eliminatórios e critérios.**

- Eliminatórios:
  - RNF-PRIV-04: isolamento por escola.
  - O ledger não conhece donos (`domain.md`).
  - Invariantes 1 e 2 verificáveis.
- Critérios:
  - Técnicos:
    - reconciliação por partição;
    - custo de conferir a partição na escrita.
  - Domínio: a partição acompanha quem guarda o dinheiro.
  - Mercado: modelo usado por ledgers de referência.

**Mercado.**

- TigerBeetle: campo `ledger` em conta e transferência; só contas do mesmo ledger transacionam diretamente. Normalmente um por moeda; em multi-tenant, um por cliente, ou por cliente e moeda.
- Formance: vários ledgers independentes por instalação.
- Square Books: tenants isolados em *shelves*.
- Padrão: partição contábil fechada, identificada por um ID opaco, sem dono de negócio.

**Opções.**

| Critério | A. Livro por escola | B. Livro único | C. Livro por cantina |
|---|---|---|---|
| Transação entre escolas | recusada pelo ledger | possível por bug | recusada |
| Reconciliação | por escola | só global | por cantina |
| Aluno com uma carteira | sim | sim | não, se a escola tiver mais de uma cantina com donos diferentes |
| Segue quem guarda o dinheiro no MVP | sim (uma cantina por escola) | não | sim |

- A:
  - Prós: isolamento e reconciliação escola a escola; casa com a RNF-PRIV-04.
  - Contras: uma coluna e uma conferência a mais em toda escrita.
  - Limitações: não serve a escola com cantinas de donos diferentes.
- B (mais simples):
  - Prós: nenhuma coluna nem conferência.
  - Contras: o isolamento depende só do código; reconciliação só global.
  - Limitações: um bug mistura escolas sem que o ledger perceba.
- C:
  - Prós: serve a cantinas de donos diferentes.
  - Contras: o aluno teria uma carteira por cantina.
  - Limitações: complexidade sem caso no MVP.

**Custo-benefício.** A custa uma coluna e uma conferência; em troca, o ledger recusa o erro mais grave de um sistema multi-escola. B economiza isso e perde a garantia. C resolve um caso que o MVP não tem.

**Riscos e limitações.** Pre-mortem ("um ano depois, A deu errado"):

- Uma escola passou a ter duas cantinas de donos diferentes. Mitigação: questão em aberto no `domain.md`; livro por cantina ou recebimento pela plataforma com repasse (UC-SIS-04).
- A conferência do livro ficou só no Go e um caminho novo a esqueceu. Mitigação: conferida também no banco.

**Recomendação:** decidido na [ADR 0005](../adr/0005-modelo-contabil-do-ledger.md): livro por escola, `book_id` em conta e transação, conferido na escrita. Regra "o livro segue quem guarda o dinheiro" e a questão das cantinas de donos diferentes registradas no `domain.md`.

**O que mudaria a recomendação:** cantinas de donos diferentes na mesma escola (C, ou A com recebimento pela plataforma); plataforma recebendo e repassando (v2), que cria um livro da plataforma.

## 2. Concorrência no saldo

**Pergunta.** Como garantir que débitos simultâneos na mesma conta nunca deixem o saldo abaixo do permitido, sem perder atualização? Errar aqui é o pior defeito possível do ledger: saldo negativo ou crédito sumido, e com histórico permanente não há como apagar o erro, só compensar. Reverter a técnica depois é barato no banco (é código do adaptador), mas caro em confiança se o bug chegar a produção.

**Eliminatórios e critérios.**

- Eliminatórios:
  - Invariante 4 do `domain.md`: conta que não pode ficar negativa nunca fica negativa; conta bloqueada para débito não é debitada.
  - RNF-CARGA-03: 20 compras simultâneas na mesma carteira; todas respondem, aprovadas ou recusadas, **sem erro**.
  - RNF-CARGA-02: p99 da autorização < 200 ms a 100 compras/s.
  - RNF-AUD-06 e ADR 0004: débito, compra, auditoria e outbox na mesma transação do Postgres.
- Critérios:
  - Técnicos:
    - corretude evidente no código;
    - tentativas desperdiçadas e latência sob disputa;
    - risco de deadlock.
  - Domínio: recusa por saldo como resultado de negócio, não como erro.
  - Mercado: adoção em ledgers de referência.

**Mercado.**

| Técnica | Como | Quem usa |
|---|---|---|
| Condição de saldo (*balance locking*) | a escrita só acontece se o saldo resultante ficar na faixa exigida (ex.: `>= 0`) | Modern Treasury (recomendada), TigerBeetle (flags `debits_must_not_exceed_credits`) |
| Lock pessimista | `SELECT ... FOR UPDATE` em todas as contas, ordenadas por ID, antes de conferir e gravar | pgledger (`pgledger_create_transfers`: "sort them to prevent deadlocks", depois `FOR UPDATE` uma a uma) |
| Versão otimista | quem chama manda a versão lida; grava só se não mudou; senão falha e repete | Modern Treasury (`lock_version`, opcional) |
| `SERIALIZABLE` | lê, decide e grava normalmente; o Postgres aborta uma das transações em conflito (`40001`) e a aplicação repete a transação inteira | documentação do Postgres |

- Fato documentado (Modern Treasury, *How to scale a ledger, part IV*):
  - Com R$ 1.000 na conta e duas transações simultâneas de R$ 250 e R$ 750, a versão otimista custa seis chamadas (lê, tenta, falha, relê); a condição de saldo, duas.
  - Em conta quente, a versão sobe tão rápido que "some Transactions will never be able to commit".
  - Por isso recomendam a condição de saldo: ela expressa a intenção real ("não deixar negativo"), não "nada mudou".
- Fato documentado (Postgres, *Transaction Isolation*): em `READ COMMITTED`, o `UPDATE` que encontra a linha alterada por outra transação espera o commit e reavalia o `WHERE` na versão nova. Em `SERIALIZABLE`/`REPEATABLE READ`, a transação falha com `could not serialize access due to concurrent update`, e a orientação é "retry the whole transaction from the beginning". A mesma página avisa que `READ COMMITTED` vê um *snapshot* inconsistente em comandos com condições complexas sobre outras linhas.
- Medição publicada (pgledger, M3 MacBook Air, Postgres 17.5, 20 workers): 10.637 transferências/s com 50 contas; 7.559/s com 10 contas, queda de ~29% causada pela espera de lock. Mostra o custo da disputa, mas não compara técnicas.
- Evidência fraca:
  - A afirmação anterior deste discovery de que o lock pessimista é "cerca de 2× mais rápido que o otimista `SERIALIZABLE`" não tem fonte verificada e foi retirada.
  - O artigo de Ports e Grittner sobre o SSI (implementação do `SERIALIZABLE` no Postgres) mede, num benchmark com leituras e escritas misturadas, vazão do `SERIALIZABLE` próxima à do *snapshot isolation* e cerca do dobro da do lock em duas fases (422 contra 208 req/s). Não mede uma linha muito disputada, que é o caso da carteira na RNF-CARGA-03; ali a taxa de repetição do otimista cresce, o que é inferência, não medição.
- Anti-padrão (Fluxo-De-Caixa): lock distribuído no Redis (Redlock) nas duas contas, saldo lido, somado em memória e gravado pelo ORM.
  - O lock é liberado ao sair do método que o adquire (`await using`), antes da gravação: não protege nada, e o saldo sofre *lost update*.
  - Lock externo com expiração e sem *fencing token* não garante exclusão: uma pausa de GC ou de rede maior que a expiração deixa dois processos gravando.
  - Contas travadas na ordem origem → destino: A→B e B→A simultâneos esperam um pelo outro até o timeout.

**Opções.** Eliminadas antes da comparação:

- Lock distribuído fora do banco: não participa da transação (anti-padrão acima).
- Fila com escritor único por conta: a compra deixaria de ser autorizada de forma síncrona na mesma transação da auditoria e do outbox.

| Critério | A. Update condicional (`READ COMMITTED`) | B. `FOR UPDATE` ordenado + checagem em Go | C. Versão otimista | D. `SERIALIZABLE` + repetição |
|---|---|---|---|---|
| Corretude evidente | a regra está no `WHERE`; um comando | a regra está no Go, depois do lock; lock esquecido = bug silencioso | a regra está no Go; versão no `WHERE` | qualquer código fica correto, desde que repita |
| Tentativas desperdiçadas na disputa | nenhuma: espera e reavalia | nenhuma: espera | muitas na conta quente | muitas na conta quente |
| Erro técnico com 20 simultâneas (RNF-CARGA-03) | não | não | sim, sem repetição interna | sim (`40001`), sem repetição interna |
| Deadlock | possível sem ordem fixa | possível sem ordem fixa | não (não espera) | não por lock; aborta |
| Regra com várias linhas (saldo + limite + regras) | cada regra precisa caber num `WHERE` próprio | natural: trava e lê tudo | natural | natural |
| Adoção | Modern Treasury, TigerBeetle | pgledger | Modern Treasury (opcional) | uso geral em Postgres |

- A (mais simples):
  - Prós:
    - Um comando por conta, sem janela entre ler e gravar.
    - `RETURNING` devolve saldo e versão para o lançamento (ADR 0005).
    - Recusa por saldo é "0 linhas", não exceção.
  - Contras: 0 linhas não diz o motivo (saldo, bloqueio ou conta inexistente) e exige uma leitura extra só no caminho de recusa.
  - Limitações: serve a regras que cabem numa condição sobre a própria linha.
- B:
  - Prós:
    - Explícito ("trava, confere, grava").
    - Acomoda regras que dependem de várias linhas lidas juntas.
  - Contras:
    - Dois comandos por conta.
    - A garantia depende de nunca esquecer o `FOR UPDATE`.
  - Limitações: as mesmas de disputa de A.
- C:
  - Prós:
    - Não segura lock entre a leitura e a escrita.
    - Serve a "o cliente viu o saldo X e confirma sobre X".
  - Contras:
    - Na conta quente, repetição sem fim.
    - Quem chama precisa repetir.
  - Limitações: protege contra mudança, não contra saldo negativo.
- D:
  - Prós:
    - O código fica simples como se não houvesse concorrência.
    - Protege também regras entre várias tabelas.
  - Contras:
    - Toda transação precisa de laço de repetição.
    - Sob disputa, a taxa de aborto cresce e a latência fica imprevisível.
  - Limitações: o custo de detecção de conflito existe em toda transação, não só nas disputadas.

**Custo-benefício.**

- A: um `UPDATE` com `WHERE` e uma leitura extra só na recusa.
- B: um `SELECT ... FOR UPDATE` a mais por conta, e disciplina.
- C e D: um laço de repetição em todo chamador e latência variável no recreio, exatamente o cenário da RNF-CARGA-03.

No volume do projeto (~1 compra/s por cantina), A e B têm o mesmo desempenho; A ganha em corretude evidente. Reverter de A para B ou D é trocar código do adaptador, sem migração.

**Riscos e limitações.** Pre-mortem ("um ano depois, A deu errado"):

- Alguém acrescentou uma checagem prévia em Go (`SELECT` do saldo, depois `UPDATE` sem condição) e reabriu a janela de *lost update*. Mitigação: o teste da RNF-CARGA-03 roda em todo `make check`; a condição de saldo fica só no `WHERE`.
- O bloqueio para débito foi checado fora do `UPDATE` e uma compra passou durante o bloqueio. Mitigação: `AND NOT debit_blocked` na mesma condição.
- Deadlock entre estorno (receita → carteira) e compra (carteira → receita) simultâneos. Mitigação: contas sempre atualizadas em ordem de ID; teste com movimentos nos dois sentidos.
- Transação longa (chamada externa dentro dela) segurou o lock da carteira e estourou o p99. Mitigação: nada de I/O externo dentro da transação (já exigido pela RNF-DISP-01); `lock_timeout` (pergunta 12).
- Limite diário (spec 0003) é outra linha disputada com a mesma natureza; se tratado com leitura prévia, repete o primeiro risco. Mitigação: mesma técnica (update condicional no gasto do dia).

**Recomendação:** A, update condicional em `READ COMMITTED`, com bloqueio para débito na mesma condição, contas em ordem de ID e `CHECK` no banco como segunda barreira. É a técnica de referência de mercado para "não deixar negativo", sem tentativas desperdiçadas na disputa do recreio. A coluna `version` permite C no futuro sem mudar o schema.

**O que mudaria a recomendação:**

- Autorização passar a depender de regras que não cabem numa condição de linha (ex.: várias contas ou tabelas lidas juntas para decidir): B.
- Cliente precisar confirmar sobre um saldo exibido ("você viu R$ 70, confirma?"): C como opção de quem chama, como no Modern Treasury.
- Disputa praticamente inexistente e regras espalhadas por muitas tabelas: D, aceitando o laço de repetição.
- Volume de milhares de débitos por segundo na mesma conta: escritor único em lote (modelo TigerBeetle), revendo a autorização síncrona.

```sql
UPDATE ledger.accounts
SET balance = balance + $delta, version = version + 1
WHERE id = $id
  AND NOT ($delta < 0 AND debit_blocked)
  AND (allow_negative OR balance + $delta >= 0)
RETURNING balance, version;
-- $delta já calculado pelo lado normal da conta (ADR 0005)
-- 0 linhas: recusa; uma leitura distingue saldo, bloqueio e conta inexistente
```

## 3. Contas quentes

**Pergunta.** As contas compartilhadas, receita da cantina e entrada de Pix, entram em quase toda transação: toda compra credita a receita, toda recarga debita a entrada de Pix. Com saldo materializado, cada uma vira uma linha disputada, e as transações de alunos diferentes passam a esperar umas pelas outras. Como evitar que isso estoure a latência? Errar para menos: fila no recreio e p99 acima da meta. Errar para mais: complexidade (assincronia, subcontas) sem retorno. Reverter é moderado: mudar como o saldo de uma conta é mantido afeta o schema e a ADR 0005 (saldo resultante e versão em cada lançamento).

**Eliminatórios e critérios.**

- Eliminatórios:
  - RNF-CARGA-01 e RNF-CARGA-02: 100 compras/s com p99 < 200 ms.
  - RNF-AUD-06 e ADR 0004: compra autorizada de forma síncrona, na mesma transação da auditoria e do outbox.
  - ADR 0005: saldo resultante e versão em cada lançamento; invariante "saldo materializado = soma dos lançamentos".
- Critérios:
  - Técnicos:
    - tempo de espera na linha disputada;
    - complexidade de escrita e de leitura do saldo;
    - impacto no modelo da ADR 0005.
  - Domínio: saldo da receita e da entrada de Pix legível para o relatório da cantina.
  - Mercado: como ledgers de referência tratam a conta quente.

**Mercado.**

- Modern Treasury (*Behind the scenes: how we built Ledgers for high throughput*, e doc *Design a Ledger for Concurrency*):
  - Descreve o problema como "inherent to double-entry accounting at scale": uma conta comum presente na maioria das transações, como a conta de liquidação de um cartão de débito.
  - Solução: API híbrida. Lançamento com trava de saldo, ou que pede o saldo resultante, é processado de forma síncrona, com lock. Os demais vão para uma fila e são aplicados em lote; meta de 60 s, p90 medido de 1 s.
  - Recomendação da doc: "the hot account receives only asynchronous entries"; travar só as contas do usuário, não a compartilhada.
  - Custo declarado: lançamento assíncrono não devolve o saldo resultante na hora.
  - Escala medida: 1.200 transações/s (4.800 lançamentos/s).
- Subcontas (*sharding*): a conta lógica vira N linhas; cada transação escolhe uma por hash ou rodízio; o saldo é a soma (artigo de arquitetura do Google Cloud sobre *hot rows* em pagamentos).
- TigerBeetle: escritor único com transferências em lote; não há disputa de lock, mas a escrita deixa de ser parte da transação do chamador.
- Medição publicada (pgledger, notebook M3, 20 workers): 2,6 ms por transferência com 10 contas disputadas, 7.559 transferências/s.

**Estimativa para nós** (a medir na Fase 6):

- Tempo de lock na linha da receita: do `UPDATE` até o commit. Ordem de grandeza: poucos ms (2,6 ms por transferência inteira no pgledger, num notebook).
- Taxa por conta de receita: ~1 compra/s por cantina, porque as 100 compras/s da RNF-CARGA-01 vêm de ~100 escolas.
- Ocupação da linha = taxa × tempo de lock:
  - caso real: 1/s × 3 ms ≈ 0,3%, espera desprezível;
  - pior caso, as 100 compras/s numa só cantina: 100/s × 3 ms ≈ 30%, espera média da ordem de 1 ms (aproximação de fila simples);
  - saturação da linha: ~300 compras/s numa só conta.
- Conclusão provisória: contenção não é risco no volume do projeto. A estimativa usa medição de outro sistema e uma aproximação; precisa de confirmação no teste de carga.

**Opções.** Eliminada antes da comparação: escritor único em lote (modelo TigerBeetle), pelo mesmo motivo da pergunta 2.

| Critério | A. Tudo materializado e síncrono | B. Saldo da conta quente aplicado de forma assíncrona | C. Conta quente sem saldo materializado | D. Subcontas |
|---|---|---|---|---|
| Espera na conta quente | ocupação × tempo de lock; desprezível no volume real | nenhuma | nenhuma | dividida por N |
| Saldo resultante e versão no lançamento (ADR 0005) | sim, em todas as contas | não na hora; exige estrutura separada, porque o lançamento é imutável | não | sim, mas da subconta, não da conta |
| Leitura do saldo da conta quente | uma linha | uma linha, atrasada | soma dos lançamentos ou snapshot | soma de N linhas |
| Peças novas | nenhuma | fila, job em lote, reconciliação do atraso | consulta de soma ou snapshot periódico | roteamento e agregação |
| Mercado | pgledger, padrão em volume baixo | Modern Treasury | variante simples de B | sistemas de pagamento de alto volume |

- A (mais simples):
  - Prós:
    - Um modelo só para todas as contas; ADR 0005 intacta.
    - Saldo da receita sempre exato e legível.
  - Contras: toda compra da cantina passa pela mesma linha.
  - Limitações: deixa de servir perto da saturação da linha (centenas de compras/s numa só conta).
- B:
  - Prós:
    - Elimina a disputa e mantém saldo materializado.
    - Validado em escala pelo Modern Treasury.
  - Contras:
    - Fila, job e janela em que o saldo da receita está atrasado.
    - O saldo resultante não cabe no lançamento imutável; precisa de outra tabela.
  - Limitações: só para contas que nunca precisam checar saldo na escrita.
- C:
  - Prós:
    - Elimina a disputa sem fila nem job.
    - Escrita da compra fica com uma linha travada só, a carteira.
  - Contras:
    - Muda a ADR 0005 para essas contas: sem saldo resultante nem versão no lançamento.
    - Leitura do saldo da receita fica cara sem snapshot.
  - Limitações: só para contas que nunca precisam checar saldo na escrita.
- D:
  - Prós:
    - Mantém tudo síncrono e com saldo resultante.
    - Escala quase linear com N.
  - Contras:
    - A conta lógica deixa de ser uma linha; reconciliação e extrato agregam N.
    - Escolha de N e roteamento viram configuração a manter.
  - Limitações: dilui a disputa, não a elimina.

**Ordem dentro da transação.** A versão anterior deste discovery recomendava gravar os lançamentos primeiro e atualizar os saldos por último. Isso não funciona com a ADR 0005: o saldo resultante do lançamento só existe depois do `UPDATE`. Dentro do ledger, a ordem é `UPDATE` do saldo e depois `INSERT` do lançamento. O que define o tempo de lock é a posição do débito na transação do canteen:

- Débito no fim da transação: lock mais curto, mas a compra já gravada como aprovada precisaria ser desfeita (savepoint) quando o saldo for insuficiente.
- Débito na ordem natural do fluxo de compra (passo 6, antes de gravar compra, auditoria e outbox): o lock dura alguns `INSERT` a mais, da ordem de 1 ms.

Na ocupação estimada, o ganho de mover o débito para o fim é desprezível e custa um savepoint em todo caminho de recusa. Fica a ordem natural.

**Custo-benefício.**

- A: custo zero além do que já existe; benefício suficiente com folga de duas ordens de grandeza no volume real.
- B: fila, job, tabela de saldos resultantes e monitoramento do atraso, para resolver um problema que não aparece abaixo de centenas de compras/s por conta.
- C: barato de construir, mas abre exceção na ADR 0005 e encarece o relatório da cantina.
- D: complexidade permanente de agregação para diluir uma disputa que não existe no volume do projeto.

**Riscos e limitações.** Pre-mortem ("um ano depois, A deu errado"):

- O teste de carga concentrou as 100 compras/s numa cantina e o p99 subiu. Mitigação: o cenário da Fase 6 mede os dois casos (distribuído e concentrado); a estimativa diz que mesmo o concentrado cabe.
- Uma transação de compra ficou lenta por outro motivo (consulta de regras pesada, I/O externo) e segurou o lock da receita. Mitigação: `lock_timeout` (pergunta 12), nada de I/O externo na transação, e o débito na ordem natural, depois das verificações.
- A v2 com recebimento pela plataforma e repasse (UC-SIS-04) criaria uma conta de entrada de Pix da plataforma inteira, quente de verdade. Mitigação: reavaliar esta pergunta quando a v2 entrar; recargas são menos frequentes que compras.
- Um relatório somando os lançamentos da receita em tempo real ficou lento. Mitigação: com A, o relatório lê o saldo materializado.

**Recomendação:** A. Materializar o saldo de todas as contas, de forma síncrona, com o débito na ordem natural do fluxo de compra. No volume do projeto a linha da receita fica ocupada menos de 1% do tempo; qualquer alternativa custa complexidade ou uma exceção na ADR 0005 sem retorno. Medir na Fase 6 com cenário concentrado.

**O que mudaria a recomendação:**

- Medição na Fase 6 com p99 acima da meta por espera na receita ou na entrada de Pix: C para a conta afetada, revendo a ADR 0005 para contas sem saldo resultante; D se o saldo resultante dessas contas for indispensável.
- Centenas de compras/s numa só conta (cantina muito maior, ou conta da plataforma na v2): B, no modelo híbrido do Modern Treasury.
- Saldo da receita passar a ser condição de escrita (ex.: limite de receita): continua A, porque B e C deixam de ser possíveis.

## 4. Duas fases (pendente → lançada ou anulada)

**Pergunta.** O ledger precisa de transação em duas fases, que reserva o valor e depois lança, anula ou expira? Errar para menos custa uma migration (status na transação, saldo pendente na conta) quando surgir o primeiro caso. Errar para mais custa estados, saldos e expiração sem uso.

**Eliminatórios e critérios.**

- Eliminatórios: compra atômica, aprovada ou recusada inteira (`domain.md`); crédito da recarga só depois do Pix confirmado.
- Critérios:
  - Técnicos: estados e saldos a manter.
  - Domínio: existência de reserva real no MVP.
  - Mercado: quando as referências usam duas fases.

**Mercado.**

- TigerBeetle: transferência pendente reserva o valor (`debits_pending`); depois é lançada (total ou parcial), anulada ou expira.
- Modern Treasury: transação nasce `pending` e passa a `posted` quando o dinheiro se move no banco.
- Uso típico: autorização de cartão, escrow, pagamento que depende de sistema externo.

**Opções.**

- A. Sem duas fases (mais simples):
  - Prós: um estado e um saldo.
  - Contras: reserva futura exige migration.
  - Limitações: não serve a encomenda antecipada.
- B. Duas fases desde já:
  - Prós: pronto para reservas.
  - Contras: status, saldo pendente e expiração sem nenhum caso no MVP.
  - Limitações: nenhuma relevante.

Casos da cantina:

- Compra: autorização e captura no mesmo instante, na mesma transação. Não há reserva.
- Recarga: o crédito só nasce quando o Pix é confirmado. A cobrança pendente vive no payments e no pedido de recarga do canteen, não no ledger.
- Primeiro caso real: encomenda antecipada (UC-RESP-12, v2), que reservaria o valor até a retirada.

**Custo-benefício.** A não custa nada agora; adicionar depois é migration e código, sem refazer o modelo da ADR 0005. B paga hoje por um caso da v2.

**Riscos e limitações.** Pre-mortem ("um ano depois, A deu errado"): a encomenda antecipada entrou e exigiu migrar transações já gravadas. Mitigação: transações existentes viram "lançadas" por padrão; a migration só acrescenta colunas.

**Recomendação:** A. Sem duas fases no MVP; registrar como fora de escopo da spec 0001.

**O que mudaria a recomendação:** encomenda antecipada (UC-RESP-12) ou qualquer reserva de saldo antes da retirada.

## 5. Saldos

**Pergunta.** Quantos saldos cada conta tem, e como o saldo é lido a partir dos lançamentos? Errar é caro: a leitura do saldo define o significado de todo número gravado, e o histórico é permanente.

**Eliminatórios e critérios.**

- Eliminatórios:
  - Invariantes 3 e 4: saldo materializado = soma dos lançamentos; conta que não pode ficar negativa nunca fica.
  - Pergunta 4: sem duas fases no MVP.
- Critérios:
  - Técnicos:
    - custo das somas;
    - facilidade de detectar erro pelo saldo.
  - Domínio: saldo legível para aluno, responsável e cantina.
  - Mercado: modelo da contabilidade e dos ledgers de referência.

**Mercado.**

- Quantidade de saldos:
  - Modern Treasury expõe três: lançado (*posted*), pendente (lançado + pendente) e disponível (entradas lançadas − saídas lançadas e pendentes).
  - TigerBeetle guarda quatro acumulados (débitos e créditos, pendentes e lançados), e o saldo sai deles.
- Leitura do saldo:

| Forma | Como | Quem usa |
|---|---|---|
| Convenção única | saldo = créditos − débitos em toda conta; a conta de origem do dinheiro fica negativa | Formance (`@world` negativa), pgledger |
| Lado normal | cada conta tem `normal_balance` (`debit` ou `credit`); lançamento na direção igual soma, na oposta subtrai | contabilidade, Modern Treasury, simple-ledger |
| Acumulados separados | débitos e créditos guardados à parte; quem lê decide o lado | TigerBeetle (flags `debits_must_not_exceed_credits` e `credits_must_not_exceed_debits`) |

**Opções.**

- Quantidade: um saldo (lançado) ou três (lançado, pendente, disponível). Sem duas fases, os três coincidem; ficar com um.
- Leitura: as três formas da tabela acima, comparadas na ADR 0005.

Na cantina, o lado sai do que a conta é para a cantina, dona do livro:

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

Conferência a qualquer momento: entrada de Pix = carteiras + receita. Na convenção única, a entrada de Pix apareceria −120: mesma informação, leitura contraintuitiva. A dúvida apareceu na própria discussão: "entrou Pix, não deveria ser crédito?".

**Custo-benefício.**

- Lado normal:
  - Custo: uma coluna na conta e um `CASE` nas somas.
  - Benefício: toda conta saudável fica positiva, e **negativo é sempre anomalia**. Uma checagem só ("nenhuma conta negativa") serve à reconciliação e a alertas, e `allow_negative` vira exceção rara em vez de flag de toda conta de sistema.
- Convenção única: soma direta, mas leitura que depende de conhecer a convenção.
- Acumulados separados: flexíveis, mas sem saldo pronto para o update condicional (pergunta 2).

**Riscos e limitações.** Pre-mortem ("um ano depois, o lado normal deu errado"):

- Uma conta foi aberta com o lado errado. Mitigação: tabela fixa no canteen coberta por teste; o `CHECK` de saldo recusa o primeiro lançamento.
- Uma consulta somou lançamentos sem o `CASE`. Mitigação: a reconciliação compara com o saldo materializado.

**Recomendação:** decidido na [ADR 0005](../adr/0005-modelo-contabil-do-ledger.md): um saldo só (lançado), lido pelo lado normal da conta; `normal_balance` definido na abertura por quem chama; lançamento com direção e valor positivo. Invariante 2 passou a "por livro, total de débitos = total de créditos".

**O que mudaria a recomendação:** duas fases (pergunta 4) traz saldo pendente e disponível; plano de contas completo continua no lado normal.

## 6. Fronteira de dados

**Pergunta.** O que o ledger guarda e o que fica no canteen? Errar para mais acopla o ledger ao negócio e duplica dados; errar para menos obriga a consultar o canteen para ler o próprio extrato contábil.

**Eliminatórios e critérios.**

- Eliminatórios: o ledger não conhece donos; FK só dentro do schema `ledger` (`CLAUDE.md`, `domain.md`).
- Critérios:
  - Técnicos:
    - reconciliação legível só com dados do ledger;
    - nenhum dado duplicado entre contextos.
  - Domínio: detalhe de negócio (itens, operador, regras) com quem decide sobre ele.
  - Mercado: separação usada pelas referências.

**Mercado.**

- TigerBeetle: só o contábil, mais campos opacos para ligar ao sistema de origem:
  - `user_data_128`: quem ou o quê;
  - `user_data_64`: segundo carimbo de tempo;
  - `user_data_32`: onde;
  - `code`: por quê (tipo de conta ou motivo da transferência).
- Formance e pgledger: metadados livres (JSON).
- Stripe e Square: o ledger é registro de fatos financeiros; o detalhe de negócio fica nos sistemas de origem.

**Opções.**

- A. Só o contábil mais código do motivo (modelo TigerBeetle):
  - Prós: ledger pequeno e estável; extrato contábil legível sem o canteen.
  - Contras: o detalhe exige juntar com o canteen pelo ID da transação.
  - Limitações: nenhum campo livre para necessidades futuras.
- B. Só o contábil, sem motivo (mais simples):
  - Prós: o mínimo possível.
  - Contras: reconciliação e extrato precisam do canteen para saber o que cada transação é.
  - Limitações: o ledger sozinho não explica nada.
- C. Contábil mais metadados livres (JSON):
  - Prós: flexível.
  - Contras: negócio vaza para o ledger; sem schema, sem garantia.
  - Limitações: dado duplicado com o canteen.

| No ledger | No canteen |
|---|---|
| livro, contas, lado normal, se pode ficar negativa, bloqueio para débito | de quem é cada conta (aluno, cantina) |
| transação: ID, código do motivo (`purchase`, `refund`, `top_up`), transação que ela reverte, instante | compra: itens, preços, categorias, versão das regras, operador, terminal |
| lançamentos: conta, direção, valor, saldo resultante, versão da conta | auditoria, eventos, avisos |

**Custo-benefício.** A custa uma coluna (`code`) e dá extrato e reconciliação contábeis autônomos. C economiza schema hoje e cobra em acoplamento e duplicação.

**Riscos e limitações.** Pre-mortem ("um ano depois, A deu errado"): um relatório contábil precisou de um dado de negócio a cada linha e ficou lento por juntar com o canteen. Mitigação: o relatório é do canteen, que já tem o detalhe e o ID da transação.

**Recomendação:** A. O canteen guarda o ID da transação do ledger; o código do motivo fica no ledger.

**O que mudaria a recomendação:** outro consumidor do ledger, fora do canteen, que precise de contexto sem acesso a ele (ex.: ledger extraído para serviço próprio).

## 7. Idempotência

**Pergunta.** Como garantir que repetir a mesma operação não gere duas transações? Errar aqui é crédito ou débito em dobro, permanente no histórico. Mudar depois é caro: a chave está na identidade da transação gravada.

**Eliminatórios e critérios.**

- Eliminatórios:
  - Invariante 5: uma chave de idempotência gera no máximo uma transação.
  - Recarga consumida de evento entregue pelo menos uma vez (ADR 0003, Fase 2).
  - Ledger permanente (RNF-RET-01).
- Critérios:
  - Técnicos:
    - funciona com vários processos e após restart;
    - conflito de conteúdo detectado.
  - Domínio: operação repetida legítima (mesmo aluno, mesmo item, duas vezes) não é duplicata.
  - Mercado: padrão das referências.

**Mercado.**

- TigerBeetle: o ID da transferência, gerado pelo cliente, **é** a chave de idempotência. Repetir devolve `exists`; repetir com campos diferentes devolve `exists_with_different_*` (erro). Recomenda IDs ordenáveis por tempo.
- Stripe: chave no header `Idempotency-Key`; guarda status e corpo da primeira resposta e devolve igual; parâmetros diferentes com a mesma chave é erro; chaves podem ser removidas após 24 h.
- Formance: chave de idempotência por transação.
- Anti-padrão (simple-ledger): chave = hash do conteúdo da requisição, em cache na memória do processo com TTL de 15 min.
  - Falha com vários processos ou após restart.
  - Recusa operações repetidas legítimas: o mesmo aluno comprar o mesmo item duas vezes em 15 min seria duplicata.

**Opções.**

- A. ID da transação gerado por quem chama é a chave (TigerBeetle):
  - Prós: sem tabela extra; a unicidade é a chave primária.
  - Contras: quem chama precisa gerar o ID de forma determinística na repetição (ex.: derivado da recarga).
  - Limitações: protege o ledger, não guarda a resposta da requisição.
- B. Tabela de chaves com resposta guardada (Stripe):
  - Prós: devolve a resposta original, inclusive recusa.
  - Contras: uma tabela e um fluxo a mais.
  - Limitações: é da camada da requisição, não do ledger.
- C. Hash do conteúdo (eliminada): recusa repetições legítimas.

Há duas camadas, que não se confundem:

- **Requisição** (spec 0004): o caixa reenvia a compra; o canteen devolve o resultado gravado, inclusive recusa. Opção B.
- **Ledger** (spec 0001, invariante 5): opção A. Protege também o consumidor de recarga.

**Custo-benefício.** A no ledger custa nada além da chave primária. B só onde a resposta precisa ser devolvida igual (requisição).

**Riscos e limitações.** Pre-mortem ("um ano depois, A deu errado"): o consumidor de recarga gerou um UUID novo a cada entrega do evento e creditou duas vezes. Mitigação: ID da transação derivado do ID da recarga, nunca aleatório na repetição; teste reentregando o mesmo evento.

**Recomendação:** A no ledger, com escopo global (UUID não colide entre livros) e sem expiração (ledger permanente). B na camada da requisição (spec 0004).

**O que mudaria a recomendação:** cliente externo chamando o ledger diretamente, sem canteen no meio (exigiria B no próprio ledger).

## 8. Imutabilidade e verificação

**Pergunta.** Como garantir que o histórico não seja alterado e que erros sejam detectados? Errar aqui é perder a única fonte de verdade do dinheiro. Reverter é barato no que for permissão e trigger, caro no que exigir dado já gravado (saldo resultante e versão em cada lançamento).

**Eliminatórios e critérios.**

- Eliminatórios:
  - Append-only (RNF-RET-01, `domain.md`).
  - Invariantes 1 a 6 verificáveis.
  - Pico do recreio (RNF-CARGA) sem serializar todas as escritas.
- Critérios:
  - Técnicos:
    - proteção contra bug e contra invasão pela aplicação;
    - detecção de erro;
    - custo na escrita.
  - Domínio: ameaça real de uma cantina.
  - Mercado: práticas das referências.

**Mercado.**

- Modern Treasury: versão por conta (`lock_version`) que sobe a cada lançamento; cada lançamento aponta a versão e o saldo resultante, o que permite reconstruir o saldo em qualquer momento. Distingue `effective_at` (fato econômico) de `created_at` (gravação).
- pgledger: cada lançamento guarda saldo anterior, saldo atual e versão da conta. O Fluxo-De-Caixa também grava o saldo resultante (`BalanceAfterTransaction`).
- Stripe: contas de compensação que devem zerar; o que não zera aponta erro. Prioriza detectar e delimitar o problema a impedir todo erro.
- Uber: verificação por checksum em janelas de tempo pegou bugs que lançamentos isolados não mostravam.
- Hash encadeado (cada linha guarda o hash da anterior): torna edição evidente, mas quem tem escrita total reescreve a cadeia inteira. Mais comum em trilha de auditoria do que em ledger operacional.
- SQL Server Ledger (2022+), usado pelo Fluxo-De-Caixa:
  - tabela `APPEND_ONLY`: o banco recusa `UPDATE` e `DELETE`;
  - hash encadeado por transação do banco;
  - *digest* exportado periodicamente para armazenamento imutável fora do banco, que fecha a brecha do hash encadeado;
  - irreversível depois de ligado;
  - sem equivalente nativo no Postgres; a versão manual seria hash encadeado, job de verificação e digest publicado fora;
  - o encadeamento global serializa todas as escritas (por conta reduz a disputa, mas complica a verificação).
- Anti-padrões (simple-ledger):
  - *upsert* da transação e lançamentos apagados e reinseridos; trigger só contra `DELETE`, nada contra `UPDATE`;
  - transação e lançamentos gravados sem transação do banco: uma queda no meio deixa transação desbalanceada.

Duas proteções diferentes:

- **Impedir:** permissões e triggers barram a aplicação, contra bug e invasão pela aplicação.
- **Detectar:** hash e digest externo revelam alteração feita por quem passou por cima da aplicação (DBA, superuser, acesso ao disco).

**Opções.**

- A. Permissões, trigger de partida dobrada, saldo resultante e versão, reconciliação:
  - Prós: cobre bug e invasão pela aplicação; histórico do saldo de graça; custo pequeno na escrita.
  - Contras: não detecta adulteração por quem tem acesso total ao banco.
  - Limitações: depende da reconciliação rodar.
- B. Só permissões (mais simples):
  - Prós: nada a construir além de `GRANT`.
  - Contras: transação desbalanceada por bug passa; erro só aparece em relatório.
  - Limitações: nenhuma detecção.
- C. A mais hash encadeado com digest externo:
  - Prós: detecta adulteração até por DBA.
  - Contras: serializa as escritas (ou complica a verificação, se por conta); job e armazenamento externo.
  - Limitações: sem o digest externo, não protege.

**Custo-benefício.** A custa permissões, um trigger e duas colunas por lançamento, e cobre o risco real (bug). C acrescenta custo na escrita do recreio para uma ameaça fora do cenário de uma cantina. B economiza pouco e deixa passar o bug mais grave.

**Riscos e limitações.** Pre-mortem ("um ano depois, A deu errado"):

- A role da aplicação ganhou `UPDATE` amplo numa migration. Mitigação: teste de integração confirma que `UPDATE` em lançamento falha com a role da aplicação.
- A reconciliação deixou de rodar e um erro ficou meses escondido. Mitigação: verificação de invariantes ao fim de todo teste de integração e job agendado na Fase 2 (UC-SIS-02).

**Recomendação:** A.

- Permissões: a role da aplicação tem só `SELECT` e `INSERT` em transações e lançamentos; em contas, `UPDATE` apenas nas colunas `balance`, `version` e `debit_blocked` (permissão por coluna do Postgres).
- Partida dobrada no banco: *constraint trigger* adiado (`DEFERRABLE INITIALLY DEFERRED`) confere no commit que cada transação tem débitos = créditos.
- Saldo resultante e versão em cada lançamento, com `UNIQUE (account_id, account_version)` (ADR 0005).
- Reconciliação: "saldo materializado = soma dos lançamentos pelo lado normal" por conta e "débitos = créditos" por livro; nos testes desde a Fase 1, agendada na Fase 2.
- Sem hash encadeado no MVP; candidato a v2 ("trilha com evidência de adulteração"), sempre com digest externo. `effective_at` separado só quando houver lançamento retroativo.

**O que mudaria a recomendação:** exigência regulatória ou contratual de evidência de adulteração (C); lançamento retroativo (`effective_at`).

## 9. Construir ou usar pronto

**Pergunta.** Construir o ledger sobre o Postgres do canteen ou usar um pronto? É a decisão mais estrutural: muda onde fica a garantia de atomicidade da compra.

**Eliminatórios e critérios.**

- Eliminatórios: débito, compra, auditoria e outbox na mesma transação (RNF-AUD-06, ADR 0004); compra sem dependência de serviço externo (RNF-DISP-01).
- Critérios:
  - Técnicos:
    - atomicidade sem coordenação entre sistemas;
    - peças a operar;
    - vazão necessária.
  - Domínio: o projeto existe para dominar essas decisões (`CLAUDE.md`).
  - Mercado: maturidade das opções prontas.

**Mercado.**

- TigerBeetle: banco dedicado a contabilidade, com contas, transferências, duas fases e idempotência prontos; feito para vazão muito acima da nossa.
- Formance: serviço de ledger sobre Postgres, com API e linguagem própria (Numscript).
- Modern Treasury: produto pago.

**Opções.**

- A. Módulo do canteen sobre o Postgres:
  - Prós: atomicidade com compra, auditoria e outbox na mesma transação; nenhuma peça nova.
  - Contras: invariantes, concorrência e idempotência construídas e testadas aqui.
  - Limitações: a vazão do Postgres, folgada no volume do projeto (pergunta 3).
- B. TigerBeetle:
  - Prós: corretude e vazão já provadas.
  - Contras: fora da transação do Postgres; a compra vira coordenação entre dois sistemas (duas fases ou saga); mais uma peça a operar.
  - Limitações: quebra o eliminatório de atomicidade sem coordenação.
- C. Formance ou Modern Treasury:
  - Prós: API pronta.
  - Contras: serviço externo no caminho da compra (contra a RNF-DISP-01); mesma coordenação de B.
  - Limitações: quebra os dois eliminatórios.

**Custo-benefício.** A custa construir e testar o que as opções prontas já trazem, e em troca evita coordenação entre sistemas, que é mais complexa que o próprio ledger nesse volume. B e C resolvem um problema de vazão que o projeto não tem.

**Riscos e limitações.** Pre-mortem ("um ano depois, A deu errado"): um bug de concorrência ou de invariante chegou à produção. Mitigação: teste de concorrência, verificação de invariantes e reconciliação (perguntas 8 e 15).

**Recomendação:** A, copiando das referências:

- do TigerBeetle: livro, código do motivo, flags de saldo e ID do cliente como chave;
- do Modern Treasury: condição de saldo, versão por conta, lado normal;
- do pgledger: saldo resultante e versão em cada lançamento, contas em ordem fixa.

Registrar na ADR "Ledger como módulo isolado no canteen", prevista na Fase 1.

**O que mudaria a recomendação:** vazão de milhares de transações por segundo, ou o ledger servindo vários sistemas além do canteen.

## 10. Schema

**Pergunta.** Que tabelas e colunas representam o modelo das perguntas 1, 2, 5 e 8? É a decisão mais cara de reverter do ledger: tabelas append-only com histórico permanente.

**Eliminatórios e critérios.**

- Eliminatórios:
  - ADR 0005 (modelo contábil) e invariantes 1 a 6.
  - FK só dentro do schema `ledger`, nenhuma para outros contextos (`CLAUDE.md`).
  - ADR 0002: sqlc + pgx; ADR 0001: migrations com goose.
- Critérios:
  - Técnicos:
    - invariantes garantidas pelo banco;
    - índices sem fragmentação;
    - consultas de extrato e reconciliação simples.
  - Mercado: correspondência com ledgers de referência.

**Mercado.**

- pgledger: `accounts` (saldo `numeric`, `version`, `allow_negative_balance`, `allow_positive_balance`, moeda), `transfers` (de, para, valor `> 0`) e `entries` (valor com sinal, saldo anterior e atual, versão).
- Modern Treasury: contas com `normal_balance` e `lock_version`; transações com lançamentos (`direction`, `amount` na menor unidade da moeda, `resulting_ledger_account_balances`, `ledger_account_lock_version`).
- TigerBeetle: inteiros de 128 bits e *asset scale* fixo por ledger, que avisa não dar para mudar depois; IDs gerados pelo cliente, ordenáveis por tempo.

**Opções.**

- Estrutura do movimento:
  - transferência de-para com duas pernas (pgledger `transfers`, Fluxo-De-Caixa);
  - transação com N lançamentos (Modern Treasury, contabilidade).
- Representação do lançamento: valor com sinal ou direção com valor positivo (pergunta 5 e ADR 0005).
- Identificadores:
  - `bigserial`: compacto, mas gerado pelo banco, então não serve de chave de idempotência (pergunta 7);
  - UUIDv4: gerado por quem chama, mas aleatório, fragmenta o índice B-tree;
  - UUIDv7: gerado por quem chama e ordenável por tempo.

**Custo-benefício.**

- N lançamentos custam uma tabela a mais que a transferência de-para e permitem taxa da plataforma ou pagamento com duas fontes sem mudar o schema.
- UUIDv7 custa 16 bytes por chave e serve ao mesmo tempo de chave de idempotência e de índice ordenado.

**Esboço** (detalhes na spec):

```sql
ledger.books        (id uuid PK, created_at timestamptz)
ledger.accounts     (id uuid PK, book_id FK, normal_balance direction,
                     allow_negative bool, debit_blocked bool,
                     balance bigint, version bigint, created_at,
                     CHECK (allow_negative OR balance >= 0))   -- balance pelo lado normal
ledger.transactions (id uuid PK, book_id FK, code text,
                     reverses_id uuid UNIQUE NULL FK transactions, created_at)
ledger.entries      (id bigint PK, transaction_id FK, account_id FK,
                     direction direction,                      -- enum debit | credit
                     amount bigint CHECK (amount > 0),
                     balance_after bigint, account_version bigint, created_at,
                     UNIQUE (account_id, account_version))
```

- Dinheiro em inteiro de centavos, nunca ponto flutuante nem `decimal` (simple-ledger acerta; Fluxo-De-Caixa usa `decimal`).
- `reverses_id UNIQUE`: invariante 6 garantida pelo banco.
- `debit_blocked`: bloqueio para débito conferido no mesmo `UPDATE` do saldo (pergunta 2).
- Livro da transação igual ao livro das contas: conferido na escrita (Go e trigger).
- `entries.id` em `bigint`: ordem de inserção interna; a ordem por conta vem da versão.

**Riscos e limitações.** Pre-mortem ("um ano depois, o schema deu errado"):

- Precisou-se de outra moeda. Mitigação: o expoente é fixo (ADR 0005); outra moeda vira outro livro, como no TigerBeetle.
- Uma coluna faltou num lançamento já gravado. Mitigação: a informação de negócio fica no canteen (pergunta 6); o lançamento só guarda o contábil.

**Recomendação:** decidido na [ADR 0005](../adr/0005-modelo-contabil-do-ledger.md) quanto ao modelo; o esboço acima vai para a spec 0001.

**O que mudaria a recomendação:** duas fases (status na transação e saldo pendente na conta); várias moedas no mesmo livro.

## 11. Idempotência no banco

**Pergunta.** Como implementar a idempotência do ledger (pergunta 7) no Postgres, inclusive com duas chamadas simultâneas com o mesmo ID? Errar aqui é transação duplicada ou erro técnico numa repetição legítima.

**Eliminatórios e critérios.**

- Eliminatórios: invariante 5; nenhum erro técnico em repetição legítima; mesma transação de quem chama.
- Critérios:
  - Técnicos:
    - comandos por chamada;
    - comportamento com chamadas simultâneas;
    - conflito de conteúdo detectado.
  - Mercado: padrão das referências.

**Mercado.**

- Stripe/Brandur: tabela de chaves com `UNIQUE (usuário, chave)`, parâmetros da requisição e resposta; parâmetros diferentes retornam 409; chamadas simultâneas com a mesma chave: uma trava, as outras recebem conflito.
- TigerBeetle: a chave é a chave primária; conteúdo diferente é erro.

**Opções.**

- A. `INSERT ... ON CONFLICT (id) DO NOTHING` e comparação dos campos:
  - Prós: um comando no caminho feliz; sem exceção a tratar; a segunda chamada simultânea espera o commit da primeira no índice único.
  - Contras: a comparação lê a transação existente e seus lançamentos.
  - Limitações: nenhuma relevante.
- B. `INSERT` simples e tratar `23505` (mais simples de escrever):
  - Prós: nada de `ON CONFLICT`.
  - Contras: o erro aborta a transação inteira de quem chama no Postgres; a recuperação exige savepoint.
  - Limitações: incompatível com rodar dentro da transação da compra sem savepoint.
- C. Ler antes e inserir se não existir:
  - Prós: código direto.
  - Contras: janela entre ler e inserir; com chamadas simultâneas, uma delas cai em `23505` (volta a B).
  - Limitações: não resolve a concorrência.

**Custo-benefício.** A custa uma leitura só no caminho de repetição e resolve a concorrência pelo próprio índice. B e C economizam pouco e trazem erro de transação abortada.

**Riscos e limitações.** Pre-mortem ("um ano depois, A deu errado"): a comparação deixou de conferir um campo novo e aceitou uma repetição com conteúdo diferente. Mitigação: a comparação cobre livro, código, reversão e todos os lançamentos; teste com cada campo alterado.

**Recomendação:** A.

1. `INSERT INTO ledger.transactions ... ON CONFLICT (id) DO NOTHING RETURNING id`.
2. Inseriu: segue com saldos e lançamentos.
3. Não inseriu: lê a transação existente e compara livro, código, reversão e lançamentos. Igual: devolve a existente, sem efeito. Diferente: `ErrIdempotencyConflict`.
4. Duas chamadas simultâneas com o mesmo ID: a segunda espera o commit da primeira no `INSERT` e cai no passo 3.

Sem hash do pedido: o ledger já guarda todos os campos. Hash e resposta guardada são da camada da requisição (spec 0004).

**O que mudaria a recomendação:** transação com muitos lançamentos, em que a comparação campo a campo fique cara (hash do conteúdo gravado junto).

## 12. Retentativa e limites de espera

**Pergunta.** Quando uma transação falha por motivo transitório (deadlock, conflito de serialização, espera de lock longa demais, conexão caída), quem repete, o quê, quantas vezes e com que limites de espera? Errar para menos: o caixa vê erro numa compra que daria certo (RNF-CARGA-03 exige "sem erro"). Errar para mais: repetições em cascata sobrecarregam a conta disputada e estouram a latência. Reverter é barato: é código do executor e configuração de sessão.

**Eliminatórios e critérios.**

- Eliminatórios:
  - Postgres: repetir a **transação inteira**, incluindo a lógica que decidiu o SQL; não há repetição de comando isolado.
  - RNF-CARGA-03: 20 compras simultâneas na mesma carteira respondem sem erro.
  - RNF-CARGA-02: p99 < 200 ms.
  - Idempotência de requisição (`domain.md`): o cliente repete com a mesma chave e recebe o mesmo resultado.
- Critérios:
  - Técnicos:
    - erro visível ao caixa;
    - carga extra gerada pelas repetições;
    - previsibilidade da latência;
    - risco de efeito duplicado fora do banco.
  - Domínio: recusa (saldo, limite, regra) nunca é repetida; é resultado.
  - Mercado: prática documentada pelo Postgres, Stripe e AWS.

**Mercado.**

- Postgres (*Serialization Failure Handling*):
  - `40001` (`serialization_failure`): repetir sempre.
  - `40P01` (`deadlock_detected`): recomendado repetir.
  - `23505` e `23P01`: só com cuidado, porque podem ser condição persistente.
  - "It is important to retry the complete transaction, including all logic that decides which SQL to issue", e por isso o Postgres não repete sozinho.
  - "Transaction retry does not guarantee that the retried transaction will complete; multiple retries may be needed."
- Postgres (*Client Connection Defaults*):
  - `lock_timeout`: aborta o comando que espera um lock além do limite; vale para cada tentativa de lock separadamente.
  - `statement_timeout`: aborta o comando que passa do limite. Se for menor ou igual ao `lock_timeout`, o `lock_timeout` não tem efeito.
  - `idle_in_transaction_session_timeout`: encerra sessão parada dentro de transação aberta, que seguraria locks.
  - `transaction_timeout` (Postgres 17): limita a duração da transação inteira.
  - Configurar esses limites no `postgresql.conf` "is not recommended because it would affect all sessions".
- Stripe: o cliente repete com backoff exponencial e jitter, sempre com a mesma chave de idempotência; erro de validação não é repetido.
- AWS (*Exponential Backoff and Jitter*): com 100 clientes disputando, backoff com *full jitter* reduz o número de chamadas a menos da metade em relação ao backoff sem jitter, ao custo de um pouco mais de tempo total.

**Estimativa para nós.** Na RNF-CARGA-03, as 20 compras na mesma carteira entram em fila na linha dela. Com ~3 ms por transação (ordem de grandeza da pergunta 3), a última espera ~20 × 3 ms ≈ 60 ms. O `lock_timeout` precisa ficar bem acima disso, ou o próprio teste da RNF-CARGA-03 vira erro.

**Opções.** Eliminada antes da comparação: repetir dentro do ledger, comando a comando. O ledger roda na transação de quem chama, e o Postgres exige repetir a transação inteira.

| Critério | A. Sem repetição no servidor | B. Executor repete deadlock e serialização; o resto vira 503 | C. Executor repete também espera de lock e conexão |
|---|---|---|---|
| Erro visível ao caixa num deadlock raro | sim (503) | não | não |
| Carga extra na conta disputada | nenhuma no servidor | só nos casos raros | repete justamente onde há fila |
| Latência no pior caso | `lock_timeout` | `lock_timeout` + poucos ms | até 3 × `lock_timeout` |
| Efeito duplicado fora do banco | não | não, se a transação só tocar o banco | idem; e conexão caída no commit deixa o resultado desconhecido |
| Código | nenhum | laço de repetição no executor | laço maior, mais casos |

- A (mais simples):
  - Prós: nada para construir no servidor; o cliente já repete com a mesma chave.
  - Contras: um deadlock ou conflito raro chega ao caixa como erro, contra a RNF-CARGA-03.
  - Limitações: depende de o cliente repetir bem.
- B:
  - Prós:
    - Cobre os casos que o Postgres manda repetir.
    - Fica num lugar só (o executor de transação, pergunta 14).
  - Contras: exige que a transação não tenha efeito fora do banco, ou a repetição duplica.
  - Limitações: não resolve espera longa; essa vira 503.
- C:
  - Prós: menos 503 em instabilidade de rede.
  - Contras:
    - Repetir espera de lock aumenta a fila na conta que já está disputada.
    - Repetir após conexão caída no commit repete uma transação que pode ter sido gravada (seguro só por causa da idempotência).
  - Limitações: o pior caso de latência multiplica.

**Custo-benefício.**

- A: zero de código, mas falha no eliminatório da RNF-CARGA-03 sempre que um deadlock raro acontecer.
- B: um laço pequeno no executor, testável, que atende o eliminatório.
- C: código e carga a mais para casos em que a repetição do cliente, com idempotência, já resolve sem piorar a fila.

**Riscos e limitações.** Pre-mortem ("um ano depois, B deu errado"):

- Alguém pôs um efeito fora do banco dentro da transação (publicar direto no Kafka, chamar o PSP) e a repetição duplicou. Mitigação: eventos só pelo outbox (ADR 0004); nada de I/O externo na transação (RNF-DISP-01).
- Um deadlock persistente (bug de ordem de contas) ficou escondido pelas repetições. Mitigação: métrica e log de cada repetição com o código do erro (RNF-OBS-02); o teste com movimentos nos dois sentidos (pergunta 2).
- O `lock_timeout` ficou baixo e o teste da RNF-CARGA-03 passou a dar 503. Mitigação: o valor sai da fila máxima medida, com folga, e o teste roda em todo `make check`.
- Os limites foram configurados no servidor inteiro e as migrations (goose), que precisam de locks longos, começaram a falhar. Mitigação: limites na role da aplicação ou por transação (`SET LOCAL`), nunca no `postgresql.conf`.

**Recomendação:** B.

- Executor repete a transação inteira em `40001` e `40P01`: até 3 tentativas, backoff curto com *full jitter*.
- Espera de lock acima do limite (`55P03`), `statement_timeout` e falha de conexão viram 503 com `Retry-After`; o caixa repete com a mesma chave, também com *full jitter*.
- Limites na role da aplicação, nunca globais:
  - `lock_timeout` bem acima da fila máxima da RNF-CARGA-03 (ponto de partida: 500 ms);
  - `statement_timeout` acima do `lock_timeout` (ponto de partida: 2 s);
  - `idle_in_transaction_session_timeout` para não deixar transação parada segurando lock (ponto de partida: 5 s).
- Recusa por saldo, limite ou regra, livro diferente e conflito de idempotência são resultado; nunca são repetidos.

Os valores são pontos de partida, a confirmar no teste da RNF-CARGA-03 e na Fase 6.

**O que mudaria a recomendação:**

- Adoção de `SERIALIZABLE` (pergunta 2, opção D): a repetição vira caminho normal, não rede de segurança; mais tentativas e medição da taxa de aborto.
- Cliente que não consegue repetir (terminal sem lógica de repetição): C, aceitando a carga extra.
- Medição mostrando esperas de lock perto do limite: rever a pergunta 3 antes de aumentar o `lock_timeout`.

## 13. Erros

**Pergunta.** Como o ledger comunica resultado e falha, e como isso chega ao cliente HTTP? Errar mistura recusa de negócio com erro técnico, e o caixa repete o que não devia ou desiste do que podia repetir.

**Eliminatórios e critérios.**

- Eliminatórios:
  - Convenção do projeto: erros de domínio como valores exportados do pacote do domínio, comparados com `errors.Is` e `errors.As`.
  - RNF-PRIV-05 e RNF-OBS-04: sem dado pessoal em resposta ou log; ID do trace no log.
  - O ledger não fala HTTP.
- Critérios:
  - Técnicos: o cliente decide sem ler texto.
  - Domínio: recusa de compra é resultado gravado, não falha.
  - Mercado: padrão aberto de erro HTTP.

**Mercado.**

- TigerBeetle: um código por resultado (`exceeds_credits`, `exists_with_different_amount`, `accounts_must_have_the_same_ledger`, `pending_transfer_already_posted`).
- RFC 9457: erro HTTP em `application/problem+json`, com `type` (URI estável), `title`, `status`, `detail`, `instance` e campos de extensão; não expor detalhe interno (stack, SQL).

**Opções.**

- A. Erros de domínio no ledger, traduzidos pelo canteen; HTTP em RFC 9457 com `code` estável:
  - Prós: cada camada fala a sua língua; padrão aberto no HTTP.
  - Contras: uma tradução por camada.
  - Limitações: nenhuma relevante.
- B. Erro genérico com mensagem (mais simples):
  - Prós: nada a definir.
  - Contras: o cliente precisa ler texto para decidir; recusa e falha se confundem.
  - Limitações: quebra a decisão de repetir (pergunta 12).
- C. Códigos HTTP do ledger repassados direto:
  - Prós: menos tradução.
  - Contras: o ledger passa a conhecer HTTP; detalhe interno vaza.
  - Limitações: contra o eliminatório.

**Custo-benefício.** A custa uma tabela de tradução no canteen e dá ao caixa um `code` estável para decidir. B e C economizam pouco e quebram a separação entre recusa e falha.

**Riscos e limitações.** Pre-mortem ("um ano depois, A deu errado"): um erro novo do ledger não foi traduzido e chegou ao caixa como 500. Mitigação: teste que percorre todos os erros exportados e confere a tradução.

**Recomendação:** A.

- Erros exportados no pacote do ledger: `ErrInsufficientFunds`, `ErrDebitBlocked`, `ErrAccountNotFound`, `ErrBookMismatch`, `ErrUnbalanced`, `ErrIdempotencyConflict`, `ErrAlreadyReversed`. `ErrUnbalanced` é bug de quem chama, não algo que o usuário veja.
- O canteen traduz para o seu domínio: `ErrInsufficientFunds` vira recusa da compra por saldo (gravada, com aviso); `ErrBookMismatch` vira erro interno.
- HTTP em RFC 9457, com `code` estável (ex.: `insufficient_funds`). Se a compra recusada é erro HTTP ou resultado com status `recusada` é decisão da spec 0002; ser gravada e devolvida na repetição puxa para resultado.
- Erro técnico: 500 genérico com o ID do trace; detalhe só no log.

**O que mudaria a recomendação:** API pública para terceiros, que pediria catálogo de erros versionado.

## 14. Transação compartilhada

**Pergunta.** Como canteen e ledger compartilham a mesma transação do Postgres sem que o domínio dependa de pgx? Errar é gravar fora da transação sem perceber, o que quebra a atomicidade da compra.

**Eliminatórios e critérios.**

- Eliminatórios: atomicidade da compra (RNF-AUD-06); domínio sem import de infraestrutura (`CLAUDE.md`); sqlc + pgx (ADR 0002).
- Critérios:
  - Técnicos:
    - o compilador mostra o que roda dentro da transação;
    - lugar natural para a repetição (pergunta 12).
  - Mercado: padrões usados em Go com arquitetura hexagonal.

**Mercado (Go).** sqlc gera `queries.WithTx(tx)`, que encaixa nos três padrões abaixo (artigos da Qonto e de Kaznacheev).

**Opções.**

| Critério | A. Unidade de trabalho | B. Transação na assinatura (mais simples) | C. Transação no `context` |
|---|---|---|---|
| Como | `WithinTx(ctx, func(ctx, uow) error)`; `uow` expõe as portas já ligadas à transação | método da porta recebe `pgx.Tx` | adaptador põe e tira a transação do `ctx` |
| Domínio sem pgx | sim | não | sim |
| Erro visível | compilador | compilador | nenhum: esquecer o `ctx` grava fora |

- A:
  - Prós: explícito; domínio limpo; executor é o lugar da repetição.
  - Contras: uma interface a mais.
  - Limitações: nenhuma relevante.
- B:
  - Prós: direto, sem abstração.
  - Contras: infraestrutura vaza para a porta.
  - Limitações: contra o eliminatório de domínio puro.
- C:
  - Prós: assinaturas limpas.
  - Contras: implícito; erro silencioso.
  - Limitações: atomicidade depende de disciplina.

**Custo-benefício.** A custa uma interface e dá garantia do compilador; C economiza a interface e troca por erro silencioso.

**Riscos e limitações.** Pre-mortem ("um ano depois, A deu errado"): um caso de uso chamou um repositório fora do `uow`. Mitigação: casos de uso recebem só o executor, nunca repositórios soltos; teste de atomicidade (erro no fim desfaz tudo).

**Recomendação:** A. Vai para a ADR "Portas e adaptadores com DDD tático leve", prevista na Fase 1.

**O que mudaria a recomendação:** nenhum caso de uso com mais de uma porta na mesma transação (B ou C ficariam suficientes), o que não é o caso da compra.

## 15. Testes

**Pergunta.** Como provar que o ledger está correto, inclusive sob concorrência? Errar é descobrir o bug em produção, com histórico permanente.

**Eliminatórios e critérios.**

- Eliminatórios: critérios de pronto 2 e 3 da Fase 1 (20 compras simultâneas; saldo = soma dos lançamentos); testes de integração com testcontainers (`conventions.md`).
- Critérios:
  - Técnicos:
    - cobertura de concorrência e invariantes;
    - tempo do `make check`.
  - Mercado: práticas de projetos de ledger.

**Mercado.**

- TigerBeetle: simulação determinística com falhas injetadas (VOPR).
- Projetos em Postgres: teste de estresse concorrente e conferência de invariantes ao fim.
- Testes de propriedade: sequências aleatórias de operações, conferindo invariantes após cada uma.

**Opções.**

- A. Unitários e integração por critério de aceite (mais simples):
  - Prós: diretos, ligados à spec.
  - Contras: não exercitam concorrência nem combinações.
  - Limitações: não atendem os critérios 2 e 3 da Fase 1.
- B. A mais concorrência e verificação de invariantes:
  - Prós: prova o que a Fase 1 exige; a verificação é a mesma consulta da reconciliação.
  - Contras: testes mais lentos.
  - Limitações: só os cenários escritos.
- C. B mais testes de propriedade:
  - Prós: encontra combinações não pensadas.
  - Contras: mais tempo de escrita e de execução.
  - Limitações: no banco real, lentos.
- D. Simulação determinística (VOPR): eliminada; exige controle do sistema inteiro, fora da escala do projeto.

**Custo-benefício.** B é o mínimo que atende os critérios de pronto. Propriedade no domínio puro é barata (sem banco) e acha casos de borda; no banco, fica para depois.

**Riscos e limitações.** Pre-mortem ("um ano depois, B deu errado"): o teste de concorrência passava por sorte, porque as goroutines não chegavam a disputar. Mitigação: barreira de largada (todas começam juntas) e conferência de que houve recusas por saldo.

**Recomendação:** B, mais propriedade no domínio puro.

- Domínio puro (unitário, por tabela): transação balanceada, valores, mesmo livro, reversão.
- Integração (testcontainers): cada critério de aceite contra o Postgres real, incluindo as permissões (`UPDATE` em lançamento falha com a role da aplicação).
- Concorrência: N goroutines com barreira de largada debitando a mesma carteira; ao fim, saldo nunca negativo, aprovadas = ⌊saldo inicial / valor⌋, nenhum erro técnico (RNF-CARGA-03); movimentos nos dois sentidos para deadlock (pergunta 2).
- Invariantes: função de verificação (invariantes 1 a 4 e versões sem buraco) ao fim de todo teste de integração; mesma consulta da reconciliação da Fase 2.
- Propriedade (`pgregory.net/rapid`): no domínio puro desde a primeira entrega; no banco, opcional.

**O que mudaria a recomendação:** bug de concorrência encontrado fora dos cenários escritos (propriedade no banco).

## Saídas

| Saída | O quê |
|---|---|
| [ADR 0005](../adr/0005-modelo-contabil-do-ledger.md) (aceita) | modelo contábil: livro, lado normal, direção e valor positivo, centavos, saldo resultante e versão (perguntas 1, 5 e 10) |
| ADR nova: concorrência no saldo | update condicional em `READ COMMITTED`, bloqueio para débito na condição, ordem por ID, contas quentes materializadas, repetição no executor e limites de espera (perguntas 2, 3 e 12) |
| ADR "Ledger como módulo isolado no canteen" (prevista) | construir em vez de usar pronto, fronteira de dados, o que foi copiado de cada referência (perguntas 6 e 9) |
| ADR "Portas e adaptadores com DDD tático leve" (prevista) | unidade de trabalho para transação compartilhada (pergunta 14) |
| `domain.md` (feito) | livro, lado normal, invariantes; questão em aberto: cantinas de donos diferentes na mesma escola |
| Spec 0001 | schema, idempotência, erros, testes; fora de escopo: duas fases, hash encadeado |
| Spec 0004 | idempotência da requisição: resposta guardada, chave simultânea |
| Spec 0002 | compra recusada como resultado ou erro HTTP; formato RFC 9457 |

## Fontes

- TigerBeetle: [Data modeling](https://docs.tigerbeetle.com/coding/data-modeling/), [Two-phase transfers](https://docs.tigerbeetle.com/coding/two-phase-transfers/), [Reliable transaction submission](https://docs.tigerbeetle.com/coding/reliable-transaction-submission/)
- Modern Treasury: [How to scale a ledger, part I](https://www.moderntreasury.com/journal/how-to-scale-a-ledger-part-i), [part IV](https://www.moderntreasury.com/journal/how-to-scale-a-ledger-part-iv), [Designing the Ledgers API with optimistic locking](https://www.moderntreasury.com/journal/designing-ledgers-with-optimistic-locking), [How to handle concurrent transactions](https://www.moderntreasury.com/journal/how-to-handle-concurrent-transactions), [Transaction status and balances](https://docs.moderntreasury.com/ledgers/docs/transaction-status-and-balances), [Verify prior ledger states](https://docs.moderntreasury.com/ledgers/docs/verify-prior-ledger-states), [Ledger account object](https://docs.moderntreasury.com/ledgers/reference/ledger-account-object), [Ledger entry object](https://docs.moderntreasury.com/ledgers/reference/ledger-entry-object), [Behind the scenes: how we built Ledgers for high throughput](https://www.moderntreasury.com/journal/behind-the-scenes-how-we-built-ledgers-for-high-throughput), [Design a Ledger for Concurrency](https://docs.moderntreasury.com/ledgers/docs/handle-concurrency)
- Formance: [Ledger](https://docs.formance.com/ledger)
- Square: [Books, an immutable double-entry accounting database service](https://developer.squareup.com/blog/books-an-immutable-double-entry-accounting-database-service/)
- Stripe: [Ledger](https://stripe.dev/blog/ledger-stripe-system-for-tracking-and-validating-money-movement), [Designing robust and predictable APIs with idempotency](https://stripe.com/blog/idempotency), [Idempotent requests](https://docs.stripe.com/api/idempotent_requests)
- Uber: [How LedgerStore supports trillions of indexes](https://www.uber.com/blog/how-ledgerstore-supports-trillions-of-indexes/)
- pgledger: [repositório](https://github.com/pgr0ss/pgledger) (`pgledger.sql`, função `pgledger_create_transfers`), [A ledger in PostgreSQL is fast](https://www.pgrs.net/2025/05/16/pgledger-in-postgresql-is-fast/)
- Brandur: [Implementing Stripe-like idempotency keys in Postgres](https://brandur.org/idempotency-keys)
- Postgres: [Serialization failure handling](https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html), [Transaction isolation](https://www.postgresql.org/docs/current/transaction-iso.html), [Client connection defaults](https://www.postgresql.org/docs/current/runtime-config-client.html)
- SSI no Postgres: [Ports e Grittner, Serializable Snapshot Isolation in PostgreSQL](https://arxiv.org/pdf/1208.4179)
- AWS: [Exponential backoff and jitter](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/)
- [RFC 9457: Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457.html)
- Transações em Go hexagonal: [Qonto](https://medium.com/qonto-way/transactions-in-go-hexagonal-architecture-f12c7a817a61), [Kaznacheev](https://www.kaznacheev.me/posts/en/clean-transactions-in-hexagon/)
- Contas quentes: [Hot rows, cool solutions (Google Cloud)](https://medium.com/google-cloud/hot-rows-cool-solutions-architecting-for-high-throughput-payment-systems-b0ae8bb2ec52)
- Hash encadeado: [Tamper-evident audit trails in PostgreSQL](https://appmaster.io/blog/tamper-evident-audit-trails-postgresql)
- SQL Server: [Ledger overview](https://learn.microsoft.com/sql/relational-databases/security/ledger/ledger-overview)
- Projetos de estudo (exemplos e anti-padrões, não referência): [Fluxo-De-Caixa](https://github.com/CristianoRC/Fluxo-De-Caixa) (.NET, SQL Server Ledger, Redlock), [simple-ledger-test-reference](https://github.com/gutogalego/simple-ledger-test-reference) (Node, SQLite, desafio de ledger)
