# Domínio

Cantina escolar pré-paga. Responsáveis recarregam a carteira do aluno via Pix e definem regras de consumo; o operador da cantina registra compras, que são autorizadas contra o saldo e as regras.

## Atores

| Ator | Faz login | Papel |
|---|---|---|
| Responsável | sim | adulto vinculado ao aluno; financia e controla o consumo |
| Operador da cantina | sim | atende no caixa; registra vendas |
| Admin | sim | administra escola, cantina, catálogo e usuários; papel atribuível a vários usuários, com toda ação auditada |
| Aluno | opcional, habilitado pelo responsável | dono da carteira; identificado na compra por QR ou matrícula |
| Sistema | — | jobs e consumidores de eventos, disparados por horário ou por eventos |

O que cada ator pode fazer: [use-cases.md](use-cases.md).

## Glossário

| Termo | Termo no código | Definição |
|---|---|---|
| Escola | `School` | tenant; todas as consultas são isoladas por escola; define o fuso horário dos seus "dias" |
| Cantina | `Canteen` | ponto de venda de uma escola; possui conta de receita no ledger |
| Aluno | `Student` | dono de uma carteira; guarda o ID da sua conta no ledger |
| Responsável | `Guardian` | adulto vinculado a um ou mais alunos |
| Carteira | `Wallet` | conta do aluno no ledger |
| Livro | `Book` | conjunto fechado de contas do ledger; transação só entre contas do mesmo livro; um por escola no MVP |
| Conta | `Account` | unidade contábil do ledger; tem lado normal; não conhece seu dono |
| Lado normal | `NormalBalance` | lado (débito ou crédito) em que o saldo da conta cresce |
| Transação | `Transaction` | movimento contábil com dois ou mais lançamentos; total de débitos = total de créditos |
| Lançamento | `Entry` | débito ou crédito de valor positivo em uma conta, parte de uma transação |
| Compra | `Purchase` | venda autorizada; congela preço, categoria e versão das regras |
| Estorno | `Refund` | cancelamento de compra via transação inversa |
| Recarga | `TopUp` | crédito na carteira após confirmação do Pix |
| Regras | `RuleSet` | conjunto versionado de restrições definido pelos responsáveis |
| Limite diário | `DailyLimit` | valor máximo gasto por dia no fuso da escola |
| Bloqueio | `WalletBlock` | suspensão de débitos na carteira, feita pelo responsável ou pelo aluno, com autor registrado; créditos continuam entrando |
| QR de identificação | `IdentificationToken` | token aleatório do aluno, lido no caixa; reemissão invalida o anterior |
| Consentimento | `Consent` | aceite do termo por um responsável; o aluno compra se ao menos um responsável aceitou |
| Preferências de aviso | `NotificationPreferences` | canais por tipo de aviso e limite de saldo baixo, guardados no canteen |
| Pedido de envio | `MessageRequest` | comando genérico ao notifications: destinatário, conteúdo, canais |
| Caixa de avisos | `Inbox` | avisos guardados pelo notifications, lidos pela API |
| Supervisor | `Supervisor` (papel) | usuário que autoriza estornos quando exigido |
| Permissão | `Permission` | autorização avulsa atribuída pelo admin a um usuário, além do papel (no MVP: estornar compra do mesmo dia) |
| Trilha de auditoria | `AuditLog` | registro append-only de ações de pessoa e alterações de saldo |

## Contextos

| Contexto | Tipo | Responsabilidade | Publica | Consome |
|---|---|---|---|---|
| canteen | núcleo | catálogo, alunos, regras, compras, estornos, pedidos de recarga, preferências de aviso; contém o módulo ledger | compra realizada, compra bloqueada, estorno, regras alteradas, recarga solicitada; pedidos de envio | cobrança criada, recarga confirmada |
| ledger (módulo do canteen) | núcleo | contas, transações, saldos | — | — |
| identity (módulo do canteen) | suporte | usuários, papéis, permissões, convites, vínculos responsável-aluno, consentimento | — | — |
| payments | integração | gera cobranças Pix no PSP e recebe o webhook; não conhece aluno nem vínculo | cobrança criada, recarga confirmada | recarga solicitada |
| notifications | genérico | entrega mensagens por canal (e-mail, app) e guarda a caixa de avisos; não conhece o domínio | — | pedidos de envio |
| reports (v2) | suporte | projeções de consumo | — | eventos de canteen |

Regras de fronteira:

- O ledger e o identity têm schema próprio e são acessados só pela sua porta. Sem FK entre contextos.
- Credenciais e emissão de token: no identity ou em provedor externo, decidido na ADR de autenticação (Fase 4).
- Eventos saem via outbox, na mesma transação da mudança de estado; o relay publica no Kafka.
- Entre contextos de domínio circulam só fatos, nunca comandos.
- Ao notifications vão pedidos de envio genéricos (destinatário com e-mail e ID de usuário, conteúdo, canais). Quem decide o que avisar, a quem e por qual canal é o contexto de domínio. O notifications não guarda usuários; a caixa de avisos é indexada pelo ID do usuário.

## Síncrono vs assíncrono

Critério: se pode acontecer segundos depois ou ser retentado sem problema, é candidato a evento.

| Síncrono, na transação | Assíncrono, por evento |
|---|---|
| autorização de compra (regras, limite diário, débito) | entrega de avisos |
| cadastro | crédito de recarga |
| estorno | relatórios |
| alteração de regras (nova versão) | invalidação de cache de regras |
| | estoque (v2) |

## Ledger

- **Append-only:** nada é editado ou apagado. Cancelamento é uma transação inversa que referencia a original.
- **Partida dobrada:** toda transação movimenta ao menos duas contas, e o total de débitos é igual ao total de créditos.
- **Livro:** as contas de uma escola formam um livro; transação só entre contas do mesmo livro. O livro segue quem guarda o dinheiro: no MVP, uma cantina por escola, um livro por escola.
- **Contas:**
  - uma por aluno (carteira);
  - por cantina, uma de receita e uma de entrada de Pix.
- **Não conhece donos:** aluno e cantina guardam o ID da sua conta. O ledger só conhece propriedades contábeis:
  - lado normal;
  - se pode ficar negativa;
  - ativa ou bloqueada para débito.
- **Lado normal:** definido na abertura da conta. Lançamento do mesmo lado aumenta o saldo; do lado oposto, diminui. Toda conta saudável tem saldo positivo; saldo negativo indica erro, salvo em conta marcada como podendo ficar negativa. Modelo: [ADR 0005](adr/0005-modelo-contabil-do-ledger.md).
- **Abertura de conta** é idempotente e acontece na mesma transação do cadastro do aluno.
- **Saldo materializado** atualizado com update condicional ([ADR 0006](adr/0006-concorrencia-no-saldo.md)); reconciliado com a soma dos lançamentos.
- **Idempotência** por chave única. Uma transação não pode ser estornada duas vezes.
- **Imutabilidade** garantida também por permissões do banco.
- Fechamento por operador ou turno é relatório da cantina, não conta no ledger.

### Movimentos

| Operação | Débito | Crédito |
|---|---|---|
| Recarga | entrada de Pix | carteira do aluno |
| Compra | carteira do aluno | receita da cantina |
| Estorno de compra | receita da cantina | carteira do aluno |

| Conta | O que é para a cantina | Lado normal |
|---|---|---|
| entrada de Pix | dinheiro que ela tem no banco | débito |
| carteira do aluno | crédito pré-pago que ela deve ao aluno | crédito |
| receita da cantina | o que ela ganhou com vendas | crédito |

A qualquer momento, entrada de Pix = soma das carteiras + receita.

### Recarga

- O pedido entra pelo canteen, que valida o vínculo do responsável com o aluno, se o aluno está ativo e qual cantina recebe. O canteen publica "recarga solicitada".
- O payments gera a cobrança no PSP e publica "cobrança criada", com o QR do Pix. O responsável obtém o QR consultando a recarga no canteen.
- Confirmado o pagamento, o payments publica "recarga confirmada" e o canteen lança o crédito no ledger (entrada de Pix → carteira).
- Valores mínimo e máximo de recarga definidos no cadastro da escola.
- A cobrança Pix tem prazo de expiração; depois dele, o registro da recarga passa a `expirada` e o QR não pode mais ser pago.
- Webhook atrasado de pagamento feito antes do prazo ainda credita: o pagamento foi válido, só o aviso atrasou.

### Recebimento do Pix

O recebedor do Pix é a conta bancária da própria cantina (configurada no PSP; no projeto, o PSP fake). O dinheiro real fica no banco da cantina; o ledger não guarda dinheiro, registra a quem ele pertence:

- entrada de Pix espelha o que entrou no banco;
- carteira do aluno é crédito pré-pago ainda não consumido;
- compra move crédito da carteira para a receita da cantina, sem movimentar o banco.

Não há repasse. Um modelo em que a plataforma recebe e repassa à cantina (com taxa) fica para a v2.

### Invariantes

1. Em toda transação, total de débitos = total de créditos.
2. Em cada livro, total de débitos = total de créditos.
3. Saldo materializado = soma dos lançamentos da conta, pelo lado normal.
4. Conta que não pode ficar negativa nunca fica negativa.
5. Uma chave de idempotência gera no máximo uma transação.
6. Uma transação é estornada no máximo uma vez.

## Atendimento no caixa

O caixa consulta antes o que o aluno pode comprar (cardápio filtrado pelas regras dele, saldo e limite restante) e só oferece o que é permitido. A recusa na compra fica restrita a mudanças entre a consulta e a compra (regra alterada, carteira bloqueada, saldo ou limite consumidos por outro caixa, produto esgotado). A autorização na transação é a garantia; a consulta é conveniência.

## Idempotência de requisições

Vale para toda operação que muda saldo: compra, estorno e pedido de recarga.

- O cliente (caixa ou app) gera uma chave de idempotência por operação e a envia em toda tentativa; na retentativa, a mesma chave.
- A chave é gravada com a operação, na mesma transação. Chave já existente: devolve o mesmo resultado da primeira vez (aprovada ou recusada), sem processar de novo.
- Mesma chave com conteúdo diferente é erro do cliente e é rejeitada.
- A chave é única por cantina. Recusas também guardam a chave, para que uma retentativa não mude o resultado.

## Fluxo de compra

1. Verifica a chave de idempotência; se já existe, devolve o resultado gravado.
2. Identifica o aluno (QR ou matrícula) e o operador.
3. Verifica:
   - aluno ativo;
   - ao menos um responsável com consentimento;
   - produtos disponíveis hoje e com categoria.
4. Avalia as regras do responsável (função pura no domínio), com a versão vigente conferida no banco.
5. Soma a compra ao gasto do dia, só se couber no limite diário.
6. Debita no ledger, só se a carteira não estiver bloqueada e tiver saldo.
7. Grava a compra (chave de idempotência, preço e categoria congelados, versão das regras, operador, cantina, terminal), a auditoria, o evento e o pedido de aviso no outbox.

- **Garantia na escrita:** bloqueio, limite diário e saldo são garantidos na própria escrita dos passos 5 e 6, não numa leitura anterior ([ADR 0006](adr/0006-concorrencia-no-saldo.md)). Consultar antes (ex.: o cardápio filtrado no caixa) é conveniência.
- **Compra atômica:** aprovada ou recusada inteira, nunca parcial. A recusa informa o motivo e os itens ou o valor que a causaram; o caixa ajusta e envia uma compra nova.
- **Recusa** (verificação negativa nos passos 3 a 6): grava a tentativa recusada com o motivo, a auditoria e o pedido de aviso. Nada é debitado e o gasto do dia não muda.
- **Erro técnico** em qualquer passo: desfaz tudo; nada é gravado.

## Estorno

- Só financeiro no MVP (sem estoque): transação inversa no ledger (receita da cantina → carteira do aluno), compra marcada como estornada, auditoria e aviso ao responsável.
- Limite diário: o estorno abate o gasto do dia em que a compra foi feita. Compra de hoje estornada libera o limite de hoje; compra de ontem estornada hoje não afeta o limite de hoje (o saldo volta normalmente).
- Autorização: operador com a permissão "estornar compra do mesmo dia" estorna sozinho compras do dia. Sem a permissão, ou compra de outro dia, exige supervisor.
- Prazo máximo: definido no cadastro da escola (30 dias nos dados de demo). Depois dele, nem o supervisor estorna.
- Só total: um estorno desfaz a compra inteira. Item errado: estorna a compra e registra uma nova. Estorno parcial na v2.
- Quando houver estoque (v2), o estorno publica um fato que o estoque consome.

## Responsáveis

- Um aluno pode ter vários responsáveis, todos com os mesmos direitos sobre ele; a auditoria registra quem fez cada ação.
- Regras: um conjunto por aluno, alterável por qualquer responsável; alterações simultâneas resolvidas pela versão (`If-Match`).
- Avisos: cada responsável tem as próprias preferências e o próprio limite de saldo baixo.
- Consentimento: cada responsável aceita o termo para usar o sistema; o aluno compra se ao menos um responsável aceitou (LGPD: consentimento de ao menos um dos pais ou do responsável legal).
- Bloqueio, recarga e habilitação do login do aluno: qualquer responsável. Qualquer responsável desfaz qualquer bloqueio, inclusive o de outro responsável.
- Divergência entre responsáveis fica fora do sistema.

## Regras dos responsáveis

- Conjunto de regras por aluno, versionado e imutável: cada alteração cria nova versão.
- Modos: lista de bloqueio ou lista de permissão. Alvos: produto ou categoria.
- Precedência: produto vence categoria; em empate, bloqueio vence.
- Limite diário de valor, com dia no fuso da escola.
- Concorrência otimista: versão inteira por aluno, exposta via ETag/If-Match; conflito retorna 409.
- Compra concorrente com alteração de regras: a compra usa a versão vigente no momento da sua transação e a registra. Alteração confirmada vale da compra seguinte em diante; nenhuma compra usa versão que não estava vigente.
- A compra registra a versão aplicada. Cache invalidado pelo evento de alteração, mas a autorização sempre confere a versão no banco.

## Identificação do aluno

- No cadastro, o sistema gera um token aleatório e o QR correspondente, impresso em cartão ou exibido no app do aluno.
- O QR nunca contém a matrícula ou outro dado previsível: quem conhece a matrícula não pode gerar o QR de outro aluno.
- Matrícula digitada pelo operador é a alternativa quando o aluno está sem o QR.
- Reemitir o QR gera um token novo e invalida o anterior.
- Cartão com chip (RFID/NFC) fica para a v2.

## Bloqueio de carteira

- Responsável e aluno (com login) podem bloquear a carteira.
- Cada bloqueio registra quem bloqueou. O responsável desfaz qualquer bloqueio; o aluno, só os que ele fez.
- A carteira fica bloqueada enquanto houver ao menos um bloqueio ativo.
- O bloqueio impede só débito (compra). Créditos (recarga e estorno) entram normalmente: o dinheiro já pago nunca fica preso.
- Autor e motivo do bloqueio são do canteen; o ledger só sabe se a conta está bloqueada para débito.

## Contas de usuário e e-mails

- Não há autocadastro: todo usuário nasce de cadastro feito pelo admin, porque o vínculo com aluno menor precisa ser validado pela escola.
- Admin cria usuários; cada um ativa a conta por convite enviado por e-mail (link de uso único, com validade) e define a própria senha.
- Recuperação de senha pelo mesmo mecanismo, com validade curta.
- Todo e-mail enviado ao aluno vai com cópia para o responsável. Aluno sem e-mail: vai só para o responsável. Consequência aceita: o responsável pode redefinir a senha do aluno (ele já controla o login pela habilitação).

## Avisos

- Avisos ao responsável: compra, compra recusada, estorno, recarga e saldo baixo.
- Canais: e-mail e app (caixa de avisos guardada no sistema e lida pela API). SMS e push na v2.
- O canteen decide os avisos: guarda as preferências e o limite de saldo baixo e envia pedidos de envio ao notifications. O notifications só entrega.
- O responsável escolhe, nas preferências, os canais de cada tipo de aviso.
- Preferências criadas junto com o responsável: todos os avisos por e-mail. No primeiro acesso, o sistema oferece a alteração; depois, pode alterar a qualquer momento.
- Desmarcar todos os canais de um tipo de aviso é permitido: o responsável deixa de receber aquele aviso.
- E-mails transacionais (convite, recuperação de senha) não são avisos: sempre por e-mail, fora das preferências, enviados pelo mesmo notifications.
- Saldo baixo: o limite é configurado pelo responsável nas preferências, sem valor padrão e sem obrigatoriedade. Sem limite configurado, não há aviso de saldo baixo.

## Auditoria

- Trilha única, no canteen: toda ação auditada acontece nele, inclusive os lançamentos do ledger (módulo do canteen), gravados na mesma transação que o registro de auditoria.
- O payments não tem ação de pessoa: o webhook do PSP não é auditado; o crédito resultante é auditado no canteen, com ator `sistema`.
- Se outro serviço passar a ter ação de pessoa, ele publica o fato e o canteen registra, mantendo a trilha única.
- O que é auditado e o conteúdo do registro: requisitos RNF-AUD em [non-functional-requirements.md](non-functional-requirements.md).

## Relatório de consumo

- Só o admin consulta (UC-ADMIN-06).
- Agregado: totais por produto, categoria e período, sem identificar alunos. Consumo individual só no extrato, para o responsável e o próprio aluno.

## Catálogo

- Categorias e produtos persistidos; cardápio de demonstração criado pela API de cadastro.
- Desativação em vez de exclusão, histórico de preço, flag "disponível hoje".
- Produto sem categoria não pode ser vendido.

## Multi-tenant

Uma escola e uma cantina no MVP, mas escola e cantina já existem no modelo. Toda consulta é isolada por escola.

## Fuso horário

- Instantes (quando algo aconteceu) são guardados em UTC.
- Cada escola tem o seu fuso, pelo nome IANA (ex.: `America/Manaus`), definido no cadastro. Nome, não deslocamento: as regras de horário de verão vêm da base IANA.
- Todo "dia" do domínio usa o fuso da escola: limite diário, prazo e dia da compra no estorno, relatório de consumo.

## Escopo

Casos de uso do MVP e da v2: [use-cases.md](use-cases.md).

Capacidades técnicas do MVP, que não são casos de uso: ledger, outbox, jobs, autenticação e papéis, auditoria.

## Questões em aberto

- Cantinas de donos diferentes na mesma escola: livro por cantina (o aluno teria uma carteira em cada) ou recebimento pela plataforma com repasse (UC-SIS-04, v2)?
