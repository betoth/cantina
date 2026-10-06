# UC-OPER-01. Registrar compra

- Ator: Operador da cantina
- Escopo: MVP
- Requisitos não funcionais: RNF-CARGA-01, RNF-CARGA-02, RNF-CARGA-03, RNF-PRIV-02, RNF-PRIV-04, RNF-PRIV-05, RNF-AUD-01, RNF-AUD-03, RNF-AUD-04, RNF-AUD-06, RNF-DISP-01, RNF-DISP-02

## Objetivo

Como operador da cantina, quero registrar a compra de um aluno identificado por QR ou matrícula, para que ela seja paga com o saldo da carteira dele dentro das regras definidas pelos responsáveis.

## Pré-condições

- O operador está autenticado e vinculado a uma cantina da escola.
- A cantina tem produtos ativos, com categoria.
- O aluno está cadastrado na escola da cantina, com carteira aberta.

## Fluxo principal

1. O operador identifica o aluno lendo o QR ou, sem o QR, digitando a matrícula.
2. O operador consulta o que o aluno pode comprar (UC-OPER-04) e oferece só o permitido.
3. O operador informa os itens e as quantidades e envia a compra.
4. O sistema confere que o aluno pode comprar agora: aluno ativo, consentimento de ao menos um responsável, carteira sem bloqueio, produtos disponíveis hoje, regras vigentes, limite diário e saldo.
5. O sistema debita o total da carteira e registra a compra, com preço e categoria de cada item congelados.
6. O operador vê a compra aprovada e o saldo restante do aluno.
7. Os responsáveis recebem o aviso de compra nos canais escolhidos.

## Fluxos alternativos e exceções

- **Aluno não identificado** (QR inválido ou reemitido, matrícula inexistente ou de outra escola): o operador é avisado de que o aluno não foi encontrado; nada é registrado.
- **Compra recusada** (aluno inativo, sem consentimento, carteira bloqueada, produto esgotado ou sem categoria, regra, limite diário ou saldo): a compra é recusada inteira, com o motivo e os itens ou o valor que a causaram. Nada é debitado. A tentativa fica registrada com o motivo e os responsáveis recebem o aviso de compra recusada. O operador ajusta e envia uma compra nova.
- **Retentativa** (o caixa não recebeu a resposta e reenvia a mesma compra): o sistema devolve o mesmo resultado da primeira vez, aprovada ou recusada, sem debitar de novo.
- **Compras simultâneas na mesma carteira** (outro caixa ao mesmo tempo): cada compra é autorizada contra o saldo e o limite já consumidos pelas outras; nenhuma ultrapassa.
- **Regras alteradas durante a compra:** vale a versão vigente no momento da autorização; a alteração vale da compra seguinte em diante.
- **Saldo abaixo do limite de aviso:** além do aviso de compra, o responsável que configurou o limite recebe o aviso de saldo baixo.
- **Serviço de mensageria ou de avisos fora do ar:** a compra é autorizada normalmente; os avisos chegam quando o serviço voltar.
- **Erro técnico:** nada é registrado; o operador vê o erro e pode reenviar a mesma compra.

## Resultado

- A carteira do aluno foi debitada no total da compra e a receita da cantina, creditada no mesmo valor.
- A compra está registrada com o operador, a cantina, o terminal, os itens com preço e categoria congelados e a versão das regras aplicada.
- O gasto do dia do aluno inclui a compra, para o limite diário.
- A compra consta na trilha de auditoria e no extrato do aluno.
- Os avisos aos responsáveis estão garantidos para entrega, mesmo que saiam depois.

## Regras de negócio

- [Fluxo de compra](../domain.md#fluxo-de-compra): ordem das verificações, compra atômica, recusa e erro técnico.
- [Atendimento no caixa](../domain.md#atendimento-no-caixa): consulta prévia como conveniência, autorização como garantia.
- [Idempotência de requisições](../domain.md#idempotência-de-requisições)
- [Identificação do aluno](../domain.md#identificação-do-aluno)
- [Regras dos responsáveis](../domain.md#regras-dos-responsáveis): precedência, limite diário, versão vigente.
- [Bloqueio de carteira](../domain.md#bloqueio-de-carteira)
- [Responsáveis](../domain.md#responsáveis): consentimento de ao menos um responsável.
- [Catálogo](../domain.md#catálogo): produto sem categoria não é vendido; disponível hoje.
- [Ledger](../domain.md#ledger): movimento da compra e invariantes.
- [Avisos](../domain.md#avisos): compra, compra recusada e saldo baixo.
- [Auditoria](../domain.md#auditoria)

## Specs

| Spec | O que cobre |
|---|---|
| [0001 Ledger](../specs/0001-ledger.md) | base: contas, transação em partida dobrada, débito sem saldo negativo sob concorrência; serve também ao estorno, ao extrato e à recarga |
| [0002 Compra contra saldo](../specs/0002-compra-contra-saldo.md) | aluno identificado por QR ou matrícula, produtos disponíveis e com categoria; aprovada debita, recusada por saldo fica registrada |
| [0003 Limite diário na compra](../specs/0003-limite-diario-na-compra.md) | recusa por limite, dia no fuso da escola, concorrência sobre o limite |
| [0004 Idempotência da compra](../specs/0004-idempotencia-da-compra.md) | retentativa devolve o mesmo resultado; mesma chave com conteúdo diferente é rejeitada |
| [0005 Auditoria da compra](../specs/0005-auditoria-da-compra.md) | trilha append-only, gravada na mesma transação, para compra aprovada e recusada |
| (Fase 3) | regras dos responsáveis, bloqueio de carteira e avisos na compra |
| (Fase 4) | consentimento e identidade real do operador na compra |

## Questões em aberto

- Uso de QR reemitido ou inválido não é registrado no MVP. Avaliar auditar essas tentativas como sinal de cartão perdido ou roubado.
