# Requisitos não funcionais

Metas declaradas e testáveis. Valores escolhidos para exercitar o sistema além do uso real de uma escola, não medidos em produção.

IDs no formato `RNF-GRUPO-NN`, citados no [roadmap](roadmap.md) e nas specs. A fase em que cada requisito é entregue está no roadmap.

## Carga e latência

Referência: escola de 1.000 alunos, recreio de 20 minutos, 2 a 4 caixas, menos de 1 compra por segundo. O risco real não é volume, é concorrência na mesma carteira.

| ID | Requisito |
|---|---|
| RNF-CARGA-01 | Vazão sustentada de 100 compras/s (cerca de 100 escolas em recreio ao mesmo tempo) |
| RNF-CARGA-02 | Latência da autorização de compra: p99 < 200 ms na vazão sustentada |
| RNF-CARGA-03 | 20 compras simultâneas na mesma carteira: nenhuma ultrapassa saldo ou limite diário; todas respondem (aprovada ou recusada) sem erro |
| RNF-CARGA-04 | Recarga vira saldo em < 5 s após o webhook do Pix |

## Privacidade (LGPD)

O sistema guarda dados de menores; o tratamento segue o consentimento do responsável e o melhor interesse da criança.

| ID | Requisito |
|---|---|
| RNF-PRIV-01 | Minimização: do aluno, só nome, matrícula e vínculo com o responsável. Sem CPF, foto ou data de nascimento |
| RNF-PRIV-02 | Consentimento: cada responsável aceita o termo no primeiro acesso (UC-RESP-09) e, antes disso, não usa o sistema; o aluno só compra se ao menos um responsável aceitou. Login do aluno só com habilitação do responsável (UC-RESP-08) |
| RNF-PRIV-03 | Responsável acessa só os próprios alunos |
| RNF-PRIV-04 | Isolamento por escola em toda consulta (tenant) |
| RNF-PRIV-05 | Logs e traces usam IDs, nunca nome ou matrícula |
| RNF-PRIV-06 | v2: dados pessoais de aluno desativado anonimizados após o prazo de retenção; histórico financeiro do ledger permanece |

## Auditoria

Regra geral: toda ação de pessoa e toda alteração de saldo, inclusive as feitas pelo sistema, gera registro na trilha de auditoria.

| ID | Requisito |
|---|---|
| RNF-AUD-01 | Ações de pessoa auditadas: cadastros, catálogo e disponibilidade, regras, bloqueios, habilitação de login, papéis de usuário, reemissão de QR, compras, estornos (com supervisor quando houver) e pedidos de recarga |
| RNF-AUD-02 | Alteração de saldo pelo sistema auditada: crédito de recarga, com ator `sistema` e referência ao pagamento Pix. O pagamento em si não é auditado |
| RNF-AUD-03 | Tentativas recusadas registradas com o motivo (ex.: compra bloqueada por regra ou sem saldo) |
| RNF-AUD-04 | Registro contém: quem (usuário e papel, ou `sistema`), quando, ação, resultado, entidade afetada, valores antes e depois em alterações de cadastro, origem (IP ou terminal). Compras e recargas guardam referência ao registro de origem, sem copiar valores |
| RNF-AUD-05 | Trilha append-only; a role da aplicação não tem `UPDATE` nem `DELETE` nela |
| RNF-AUD-06 | Gravada na mesma transação da ação; se a auditoria falhar, a ação falha |
| RNF-AUD-07 | Consulta só por admin (UC-ADMIN-07) |

## Disponibilidade

Sem meta de uptime percentual (projeto sem ambiente de produção).

| ID | Requisito |
|---|---|
| RNF-DISP-01 | Compra não depende de serviço externo: com Kafka, payments ou notifications fora do ar, compras continuam sendo autorizadas normalmente (saldo, limite e regras); eventos ficam no outbox até o broker voltar. Verificável com o broker parado |
| RNF-DISP-02 | Regras nunca vêm só do cache: a autorização confere a versão das regras no banco, dentro da transação. Alteração de regra vale na compra seguinte mesmo sem o evento de invalidação |

Limitação aceita: recarga paga durante a queda do broker só vira saldo quando ele voltar; até lá, compras podem ser recusadas por saldo insuficiente.

## Observabilidade

| ID | Requisito |
|---|---|
| RNF-OBS-01 | Traces OpenTelemetry de ponta a ponta, atravessando outbox e Kafka (contexto no header da mensagem) |
| RNF-OBS-02 | Métricas técnicas via OpenTelemetry: taxa, erros e latência por endpoint |
| RNF-OBS-03 | Métricas de negócio: compras aprovadas e recusadas por motivo, latência da autorização, pendentes no outbox, atraso dos consumers, divergências na reconciliação |
| RNF-OBS-04 | Logs em JSON estruturado, com ID do trace |
| RNF-OBS-05 | Dashboards Grafana versionados no repositório: recreio (compras/s, latência, recusas por motivo, erros) e saúde dos eventos (outbox, consumers, reconciliação) |

Alertas fora do MVP.

## Retenção

| ID | Requisito |
|---|---|
| RNF-RET-01 | Ledger, compras, estornos e trilha de auditoria: permanentes (append-only; registro financeiro e contábil) |
| RNF-RET-02 | Eventos publicados no outbox removidos após 7 dias |
| RNF-RET-03 | Logs e traces retidos por 14 dias |
| RNF-RET-04 | v2: dados pessoais de aluno desativado anonimizados 1 ano após a desativação, se o saldo estiver zerado (UC-SIS-05) |
