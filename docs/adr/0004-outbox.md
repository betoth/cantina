# 0004. Publicação de eventos via outbox

- Status: aceita
- Data: 2026-10-05

## Contexto e problema

Quando uma compra é gravada, o evento "compra realizada" precisa chegar ao Kafka ([0003](0003-mensageria.md)). Banco e broker são sistemas separados, sem transação comum. Gravar no banco e publicar no broker em sequência (dual write) falha de duas formas: commit sem publicação (evento perdido, responsável não é avisado) ou publicação sem commit (evento de algo que não aconteceu, ex.: recarga creditada que foi desfeita).

## Fatores de decisão

- Evento publicado se e somente se a mudança de estado foi confirmada.
- Nenhum evento perdido em queda do serviço ou do broker.
- Complexidade operacional proporcional ao projeto.

## Opções consideradas

- Publicar direto no broker após o commit
- Outbox com relay por polling
- Outbox com CDC (Debezium lendo o WAL do Postgres)

## Decisão

Escolhida: **outbox com relay por polling**.

- O evento é gravado numa tabela `outbox`, no schema do contexto, na mesma transação da mudança de estado.
- Um relay lê pendentes em ordem, publica no Kafka e marca como publicados.
- O relay pode publicar o mesmo evento mais de uma vez (queda entre publicar e marcar). Por isso os consumidores são idempotentes pelo ID do evento.

## Consequências

- Positivas: atomicidade entre estado e evento; nenhum componente extra além do próprio serviço; fácil de testar com testcontainers; migração para CDC possível depois, trocando só o relay, sem mudar a tabela `outbox` nem quem grava os eventos.
- Negativas: latência de até um intervalo de polling; carga de leitura na tabela `outbox`; limpeza periódica de eventos publicados; duplicatas tratadas nos consumidores.

## Prós e contras das opções

### Publicar direto após o commit

- Prós: simples; menor latência.
- Contras: perde eventos se o serviço cair entre commit e publicação; não há como retentar sem estado persistido.

### Outbox com polling

- Prós: atômico; só precisa do Postgres; lógica visível e testável no próprio código.
- Contras: latência de polling; carga de consulta; entrega pelo menos uma vez.

### Outbox com CDC (Debezium)

- Prós: baixa latência; sem polling; lê o WAL, sem carga de consulta; referência quando já existe Kafka Connect na infraestrutura (Outbox Event Router do Debezium).
- Contras: exige Kafka Connect e Debezium rodando e configurados; configuração de replicação lógica no Postgres; mais peças para operar e depurar.
