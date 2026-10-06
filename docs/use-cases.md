# Casos de uso

Caso de uso é o objetivo de um ator, de ponta a ponta, do ponto de vista dele. Ex.: "responsável recarrega a carteira do aluno" cobre gerar o Pix, pagar, ver o saldo e receber o aviso.

- Detalhe de um caso (fluxo principal, fluxos alternativos, resultado, specs que o implementam): arquivo `use-cases/UC-ATOR-NN-nome.md`, criado quando o caso for detalhado. Nesta lista, o ID passa a ser link para ele.
- Detalhe técnico, critérios de aceite e tarefas de implementação: specs em `specs/`, que citam os casos de uso que implementam. Um caso de uso pode ter várias specs; uma spec pode atender vários casos (ex.: a do ledger serve à compra, ao estorno e ao extrato). As tarefas viram issues no GitHub.
- Atores e termos: [domain.md](domain.md).
- IDs no formato `UC-ATOR-NN`: o prefixo indica o ator (`USR` é comum a todo ator com login). O número não indica ordem de implementação; a ordem está no [roadmap](roadmap.md).

## MVP

### Usuário

Comum a todo ator com login.

| ID | Caso de uso |
|---|---|
| UC-USR-01 | Ativar a conta pelo convite recebido por e-mail (definir senha) |
| UC-USR-02 | Recuperar senha |
| UC-USR-03 | Autenticar-se |

### Responsável

| ID | Caso de uso |
|---|---|
| UC-RESP-01 | Recarregar a carteira do aluno via Pix |
| UC-RESP-02 | Consultar saldo e extrato do aluno |
| UC-RESP-03 | Consultar cardápio (produtos, categorias e preços) |
| UC-RESP-04 | Configurar regras do aluno (bloqueio ou permissão por produto ou categoria, limite diário) |
| UC-RESP-05 | Configurar preferências de notificação (canal por tipo de aviso, limite de saldo baixo) |
| UC-RESP-06 | Receber avisos de compra, compra recusada, estorno, recarga e saldo baixo, nos canais escolhidos |
| UC-RESP-07 | Bloquear e desbloquear a carteira do aluno temporariamente (desfaz qualquer bloqueio) |
| UC-RESP-08 | Habilitar e desabilitar o login do aluno |
| UC-RESP-09 | Aceitar o termo de consentimento no primeiro acesso (até aceitar, não usa o sistema; o aluno compra se ao menos um responsável aceitou) |
| UC-RESP-15 | Consultar a caixa de avisos (canal app) |

### Operador da cantina

| ID | Caso de uso |
|---|---|
| UC-OPER-01 | Registrar compra para um aluno identificado por QR ou matrícula |
| UC-OPER-02 | Estornar compra (com supervisor quando exigido) |
| UC-OPER-03 | Marcar produto como disponível ou esgotado hoje (só a disponibilidade, sem alterar outros dados do produto) |
| UC-OPER-04 | Consultar o que o aluno pode comprar: cardápio filtrado pelas regras dele, saldo e limite restante do dia |

### Aluno

| ID | Caso de uso |
|---|---|
| UC-ALUNO-01 | Consultar saldo e extrato |
| UC-ALUNO-02 | Consultar cardápio e as próprias regras (o que pode comprar, limite restante do dia) |
| UC-ALUNO-03 | Bloquear a própria carteira e desfazer bloqueios feitos por ele |

### Admin

| ID | Caso de uso |
|---|---|
| UC-ADMIN-01 | Cadastrar escola (inclui valores mínimo e máximo de recarga e prazo máximo de estorno) e cantina |
| UC-ADMIN-02 | Cadastrar aluno (abre a carteira no ledger e gera o QR de identificação) |
| UC-ADMIN-03 | Cadastrar responsável e vincular a alunos |
| UC-ADMIN-04 | Gerenciar categorias |
| UC-ADMIN-05 | Gerenciar produtos (preço com histórico, disponível hoje, desativar) |
| UC-ADMIN-06 | Consultar relatório de consumo agregado (por produto, categoria e período, sem identificar alunos) |
| UC-ADMIN-07 | Consultar trilha de auditoria |
| UC-ADMIN-08 | Gerenciar usuários, papéis (admin, operador, supervisor) e permissões |
| UC-ADMIN-09 | Reemitir o QR do aluno, invalidando o anterior |
| UC-ADMIN-10 | Desativar aluno (deixa de comprar e sai das listas; histórico e saldo preservados) |

### Sistema

| ID | Caso de uso |
|---|---|
| UC-SIS-02 | Reconciliar saldos do ledger com a soma dos lançamentos |

## v2

### Responsável

| ID | Caso de uso |
|---|---|
| UC-RESP-10 | Registrar alergias do aluno (dado sensível, LGPD) |
| UC-RESP-11 | Limitar compras por quantidade ou horário |
| UC-RESP-12 | Encomendar lanche antecipadamente |
| UC-RESP-13 | Contestar uma compra |
| UC-RESP-14 | Exercer direitos do titular sobre os dados do aluno (consultar, corrigir, exportar, revogar consentimento) |
| UC-RESP-16 | Receber avisos por SMS ou push |

### Operador da cantina

| ID | Caso de uso |
|---|---|
| UC-OPER-05 | Fechar caixa por operador ou turno |
| UC-OPER-06 | Identificar aluno por cartão com chip (RFID/NFC) |
| UC-OPER-07 | Estornar parcialmente uma compra (itens específicos) |

### Admin

| ID | Caso de uso |
|---|---|
| UC-ADMIN-11 | Devolver saldo quando o aluno sai da escola |
| UC-ADMIN-12 | Devolver recarga ao pagador via Pix |

### Sistema

| ID | Caso de uso |
|---|---|
| UC-SIS-01 | Gerar e enviar resumo diário de consumo (opção nas preferências do responsável) |
| UC-SIS-03 | Controlar estoque por eventos |
| UC-SIS-04 | Repassar valores da plataforma à cantina, com taxa |
| UC-SIS-05 | Anonimizar dados pessoais de aluno desativado após o prazo de retenção |
