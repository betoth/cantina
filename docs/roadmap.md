# Roadmap

Plano de todas as fases: o que cada uma entrega, em que ordem e quando está pronta. A ordem segue dependência e risco: primeiro o que é mais arriscado tecnicamente (concorrência no ledger), por último o que é mais padronizado.

Casos de uso: [use-cases.md](use-cases.md). Requisitos não funcionais: [non-functional-requirements.md](non-functional-requirements.md).

| Fase | Tema | Situação |
|---|---|---|
| 0 | Harness | concluída |
| 1 | Ledger, compra e estorno | em andamento |
| 2 | Recarga: Pix fake, webhook, outbox, Kafka | planejada |
| 3 | Regras, bloqueios e avisos | planejada |
| 4 | Identidade e acesso | planejada |
| 5 | Cadastros | planejada |
| 6 | Carga e dashboards | planejada |
| 7 | Polimento | planejada |

## Fase 0: harness

Base de conhecimento e de processo para a Fase 1 começar no fluxo spec → testes → implementação → verificação → diário. Infraestrutura de código nasce na Fase 1, quando for usada.

Documentação:

- [x] Módulo Go (`go.mod`, Go 1.27)
- [x] Domínio ([domain.md](domain.md))
- [x] Arquitetura: diagrama de containers ([architecture.md](architecture.md))
- [x] Casos de uso ([use-cases.md](use-cases.md))
- [x] Requisitos não funcionais ([non-functional-requirements.md](non-functional-requirements.md))
- [x] Convenções ([conventions.md](conventions.md))
- [x] README em inglês
- [x] README em português
- [x] Diário ([journal.md](journal.md))
- [x] Roadmap

ADRs:

- [x] 0001 Migrations versionadas com goose
- [x] 0002 Acesso a dados com sqlc + pgx
- [x] 0003 Mensageria com Kafka
- [x] 0004 Publicação de eventos via outbox

Harness de IA:

- [x] `CLAUDE.md`
- [x] Skill `/spec` (com template)
- [x] Skill `/adr` (com template)
- [x] Skill `/journal`
- [x] Skill `/checkpoint`
- [x] Skill `/domain`
- [x] Skill `/use-cases`
- [x] Skill `/use-case` (com template de detalhe)
- [x] Skill `/requirements`
- [x] Skill `/conventions`
- [x] Skill `/roadmap`
- [x] Skill `/cards`: iniciar fase (milestone e issues), acompanhar, fechar fase
- [x] [Quadro no GitHub Projects](https://github.com/users/betoth/projects/1) (A fazer, Em andamento, Em revisão, Feito)

Fechamento:

- [x] Diário atualizado com o fechamento da fase

**Pronto quando:** a primeira tarefa da Fase 1 pode começar sem decisão em aberto que a bloqueie.

## Fase 1: ledger, compra e estorno

Maior risco técnico do projeto: concorrência sobre saldo e limite diário. Também nasce aqui a base transversal (auditoria, logs, traces, métricas técnicas), junto com o primeiro caso real.

Simplificações, substituídas em fases seguintes:

- Identidade fake por header (Fase 4).
- Escola, cantina, alunos com carteira, categorias, produtos e saldo inicial por seed (cadastros na Fase 5, recarga na Fase 2).
- Limite diário vindo do seed; regras de bloqueio e permissão versionadas na Fase 3.
- Sem consentimento: todos os responsáveis tratados como tendo aceitado o termo (Fase 4).
- Estorno sem supervisor (Fase 4).

Casos de uso:

- [ ] UC-OPER-01 Registrar compra
- [ ] UC-OPER-02 Estornar compra
- [ ] UC-RESP-02 Consultar saldo e extrato

Requisitos não funcionais:

- [ ] RNF-CARGA-03 20 compras simultâneas na mesma carteira
- [ ] RNF-PRIV-04 Isolamento por escola
- [ ] RNF-PRIV-05 Logs sem dado pessoal
- [ ] RNF-RET-01 Ledger, compras, estornos e trilha de auditoria permanentes
- [ ] RNF-AUD-01 Ações de pessoa auditadas
- [ ] RNF-AUD-02 Alterações de saldo pelo sistema auditadas
- [ ] RNF-AUD-03 Tentativas recusadas registradas
- [ ] RNF-AUD-04 Conteúdo do registro de auditoria
- [ ] RNF-AUD-05 Trilha append-only
- [ ] RNF-AUD-06 Auditoria na mesma transação da ação
- [ ] RNF-OBS-01 Traces
- [ ] RNF-OBS-02 Métricas técnicas
- [ ] RNF-OBS-04 Logs em JSON com ID do trace

Técnico:

- [ ] Layout `cmd/` e `internal/`
- [ ] Migrations por módulo
- [ ] `tools/go.mod` com goose e sqlc
- [ ] Makefile com `check`
- [ ] golangci-lint configurado
- [ ] `.editorconfig`
- [ ] Compose com Postgres
- [ ] `.env.example`
- [ ] CI no GitHub Actions em todo PR e push na `main`: lint, testes unitários e de integração, build
- [ ] CI confere que o código gerado pelo sqlc está atualizado
- [ ] CI roda `govulncheck`
- [ ] Proteção da `main`: merge só com CI verde
- [ ] Repositório só com squash merge e branch apagada após o merge
- [ ] Template de PR pela seção Git de `conventions.md`
- [ ] API HTTP mínima
- [ ] Coleção de requests
- [ ] Testes de integração com testcontainers
- [ ] Instrumentação com OpenTelemetry, exportando só OTLP
- [ ] OpenTelemetry Collector no compose
- [ ] Jaeger no compose (traces)
- [ ] Prometheus no compose (métricas)

ADRs:

- [ ] Ledger como módulo isolado no canteen
- [ ] Portas e adaptadores com DDD tático leve

Harness de IA:

- [ ] Skill `/tests-from-spec`
- [x] Skill `/refine` *(não planejado)* ([#3](https://github.com/betoth/cantina/issues/3))
- [x] Agent `reviewer` ([#2](https://github.com/betoth/cantina/issues/2))
- [x] Skill `/discovery` *(não planejado)* ([#7](https://github.com/betoth/cantina/issues/7))
- [ ] Agent `test-designer`
- [ ] `.claude/settings.json`

**Pronto quando:**

1. Pela coleção de requests: compra aprovada reduz o saldo; compra recusada por saldo e por limite; estorno devolve o saldo.
2. Teste com 20 compras simultâneas na mesma carteira passa.
3. Teste confirma que, após as compras concorrentes, todo saldo é igual à soma dos lançamentos da conta.
4. Compra, estorno e tentativa recusada geram registro de auditoria; um log da compra mostra o ID do trace e nenhum dado pessoal.
5. Retentativa de compra com a mesma chave de idempotência devolve o mesmo resultado e não debita de novo.

## Fase 2: recarga

Primeiro fluxo entre serviços e primeiros jobs agendados. Depende do ledger.

Casos de uso:

- [ ] UC-RESP-01 Recarregar a carteira via Pix
- [ ] UC-SIS-02 Reconciliar saldos

Requisitos não funcionais:

- [ ] RNF-CARGA-04 Recarga vira saldo em < 5 s
- [ ] RNF-DISP-01 Compra funciona com o broker parado
- [ ] RNF-RET-02 Limpeza do outbox após 7 dias

Técnico:

- [ ] Pedido de recarga no canteen (valida vínculo, aluno ativo e cantina) e consulta do status com o QR
- [ ] Serviço payments: consome "recarga solicitada", gera a cobrança e publica "cobrança criada"
- [ ] Pix fake com modos de falha: duplicado, atrasado, fora de ordem, assinatura inválida
- [ ] Webhook idempotente
- [ ] Verificação de assinatura do webhook
- [ ] Expiração da cobrança
- [ ] Webhook atrasado de pagamento feito antes da expiração ainda credita
- [ ] Outbox e relay
- [ ] Kafka no compose
- [ ] Consumer de crédito idempotente no canteen
- [ ] Agendador de jobs com lock entre instâncias (primeiro job: reconciliação; depois expiração de cobranças e limpeza do outbox)

ADRs:

- [ ] Monorepo com vários binários
- [ ] Recebedor do Pix é a conta da cantina

**Pronto quando:**

1. Pela coleção de requests: responsável pede a recarga ao canteen, obtém o QR, o Pix fake confirma e o saldo aparece em menos de 5 s.
2. Modos de falha tratados: webhook duplicado credita uma vez; atrasado de pagamento válido credita; fora de ordem não quebra; assinatura inválida é rejeitada.
3. Com o Kafka parado: compra funciona; recarga fica pendente no outbox e credita quando o Kafka volta.
4. Um trace mostra a recarga inteira, do webhook ao crédito.
5. Com duas instâncias do canteen, a reconciliação agendada roda uma vez por ciclo e não acusa divergência.

## Fase 3: regras, bloqueios e avisos

Depende da compra (regras) e dos eventos (avisos).

Casos de uso:

- [ ] UC-RESP-03 Consultar cardápio
- [ ] UC-RESP-04 Configurar regras
- [ ] UC-RESP-05 Configurar preferências de notificação
- [ ] UC-RESP-06 Receber avisos
- [ ] UC-RESP-15 Consultar a caixa de avisos
- [ ] UC-RESP-07 Bloquear e desbloquear a carteira
- [ ] UC-ALUNO-03 Bloquear a própria carteira (com identidade fake de aluno até a Fase 4)
- [ ] UC-OPER-04 Consultar se o aluno pode comprar

Requisitos não funcionais:

- [ ] RNF-DISP-02 Regras nunca vêm só do cache

Técnico:

- [ ] Serviço notifications genérico: recebe pedidos de envio e entrega por canal
- [ ] Pedidos de envio publicados pelo canteen via outbox
- [ ] Canal e-mail
- [ ] Canal app (caixa de avisos)
- [ ] E-mail fake (Mailpit) no compose
- [ ] Regras versionadas com ETag/If-Match
- [ ] Cache de regras invalidado por evento
- [ ] Bloqueio de conta para débito no ledger
- [ ] Registro de bloqueios com autor no canteen

**Pronto quando:**

1. Responsável bloqueia uma categoria; compra de produto dela é recusada, com aviso ao responsável.
2. Alteração de regra com versão desatualizada (`If-Match` antigo) retorna 409.
3. Alteração de regra vale na compra seguinte, mesmo com cache antigo.
4. Carteira bloqueada recusa compra, mas recebe recarga e estorno; aluno não desfaz bloqueio feito pelo responsável.
5. Com limite configurado, compra que deixa o saldo abaixo dele gera aviso de saldo baixo; sem limite, não gera.
6. Responsável recém-criado recebe aviso por e-mail (Mailpit); com o canal app marcado, o aviso aparece na caixa de avisos; com todos os canais de um tipo desmarcados, aquele aviso não é enviado.

## Fase 4: identidade e acesso

Troca a identidade fake pela real. Altera compra e estorno da Fase 1: compra recusada sem consentimento do responsável; estorno com supervisor quando exigido.

Simplificações, substituídas em fases seguintes:

- Escola, cantina, alunos e responsáveis continuam vindo do seed, com convite pendente (cadastros na Fase 5).

Casos de uso:

- [ ] UC-USR-01 Ativar conta pelo convite
- [ ] UC-USR-02 Recuperar senha
- [ ] UC-USR-03 Autenticar-se
- [ ] UC-ADMIN-08 Gerenciar usuários, papéis e permissões
- [ ] UC-RESP-08 Habilitar login do aluno
- [ ] UC-RESP-09 Aceitar termo de consentimento
- [ ] UC-ALUNO-01 Consultar saldo e extrato
- [ ] UC-ALUNO-02 Consultar cardápio e próprias regras
- [ ] UC-ADMIN-07 Consultar trilha de auditoria

Requisitos não funcionais:

- [ ] RNF-PRIV-02 Consentimento
- [ ] RNF-PRIV-03 Responsável acessa só os próprios alunos
- [ ] RNF-AUD-07 Consulta da trilha só por admin

Técnico:

- [ ] Módulo identity no canteen (schema próprio, acesso pela porta)
- [ ] Identidade lida do token em vez dos headers
- [ ] Papéis, permissões e autorização
- [ ] Supervisor no estorno

ADRs:

- [ ] Autenticação própria ou provedor de identidade

**Pronto quando:**

1. Admin cria um operador; ele recebe o convite no Mailpit, define a senha e faz login; compra registrada por ele aparece na auditoria com o usuário real.
2. Recuperação de senha: link expira; após a troca, sessões antigas são invalidadas; resposta igual para e-mail inexistente.
3. Responsável sem aceite não acessa o sistema; aluno sem nenhum responsável com aceite não compra; com o aceite de um deles, a compra funciona.
4. Responsável não vê aluno de outro responsável.
5. Operador sem a permissão de estorno do mesmo dia, ou estorno de compra de outro dia, é recusado sem aprovação do supervisor.
6. Aluno com login habilitado consulta o saldo; com login desabilitado, não acessa. E-mail ao aluno chega com cópia ao responsável.
7. Só admin consulta a trilha de auditoria.

## Fase 5: cadastros

Troca o seed pelos cadastros reais e remove o seed: todo dado passa pelas regras de cadastro. CRUDs mais padronizados.

Casos de uso:

- [ ] UC-ADMIN-01 Cadastrar escola e cantina
- [ ] UC-ADMIN-02 Cadastrar aluno
- [ ] UC-ADMIN-03 Cadastrar responsável
- [ ] UC-ADMIN-04 Gerenciar categorias
- [ ] UC-ADMIN-05 Gerenciar produtos
- [ ] UC-ADMIN-06 Relatório de consumo
- [ ] UC-ADMIN-09 Reemitir QR
- [ ] UC-ADMIN-10 Desativar aluno
- [ ] UC-OPER-03 Marcar disponibilidade de produto

Requisitos não funcionais:

- [ ] RNF-PRIV-01 Minimização

Técnico:

- [ ] Remoção do seed por insert direto
- [ ] Bootstrap do primeiro admin (só se não existir nenhum; e-mail por configuração)
- [ ] Script de dados de demo pela API de cadastro

**Pronto quando:**

1. Admin cadastra escola, cantina, categoria e produto; produto sem categoria não pode ser vendido.
2. Cadastro do aluno abre a carteira no ledger e gera o QR na mesma transação; se o ledger falhar, o cadastro é desfeito.
3. Cadastro do responsável cria as preferências padrão, envia o convite e vincula o aluno.
4. Reemitir o QR invalida o antigo: compra com o QR antigo é recusada, com o novo é aceita.
5. Aluno desativado não compra e some das listas; saldo e histórico continuam.
6. Operador marca produto como esgotado e a compra dele é recusada; operador não consegue alterar o preço.
7. Alteração de preço guarda histórico; compras antigas mantêm o preço congelado.
8. Relatório mostra consumo agregado por produto, categoria e período, sem identificar alunos.
9. Cadastro do aluno guarda só nome, matrícula e vínculo (RNF-PRIV-01).
10. Banco vazio + bootstrap + script de demo resulta num ambiente utilizável, sem nenhum insert direto.

## Fase 6: carga e dashboards

Mede o sistema completo contra os requisitos não funcionais. Logs, traces e métricas técnicas já existem desde a Fase 1.

Requisitos não funcionais:

- [ ] RNF-CARGA-01 100 compras/s
- [ ] RNF-CARGA-02 p99 < 200 ms
- [ ] RNF-OBS-03 Métricas de negócio
- [ ] RNF-OBS-05 Dashboards
- [ ] RNF-RET-03 Retenção de logs e traces

Técnico:

- [ ] Cenário "recreio" no k6
- [ ] Grafana no compose, lendo do Prometheus e do Jaeger
- [ ] Agregação de logs pelo Collector (ex.: Loki), consultável no Grafana
- [ ] Retenção de 14 dias configurada nos backends de logs e traces

**Pronto quando:**

1. Cenário "recreio" no k6 sustenta 100 compras/s com p99 < 200 ms.
2. Durante a carga, o dashboard do recreio mostra vazão, latência, recusas por motivo e erros.
3. Com o Kafka parado durante a carga, o dashboard de eventos mostra pendentes subindo no outbox e zerando quando ele volta.
4. Um log de compra encontrado no Grafana leva ao trace dela.

## Fase 7: polimento

- [ ] README final: arquitetura, como rodar, demo, decisões e desenvolvimento com IA, nos dois idiomas
- [ ] `make demo`: sobe o ambiente, faz bootstrap, cria dados de demo pela API e executa um roteiro com recarga, compra aprovada, compra recusada e avisos, só pelas APIs públicas
- [ ] CLI de caixa
- [ ] Avaliar tradução de `docs/` para inglês

**Pronto quando:**

1. Alguém que acabou de clonar roda `make demo`, segue o README e vê compra, recarga e aviso funcionando em menos de 10 minutos.
2. Pela CLI de caixa, o operador registra uma compra lendo o QR (ou a matrícula) e vê o resultado.
