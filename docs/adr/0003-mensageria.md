# 0003. Mensageria com Kafka

- Status: aceita
- Data: 2026-10-05

## Contexto e problema

Parte do trabalho não precisa acontecer na transação que o origina: crédito de recarga, notificações, relatórios, invalidação de cache de regras. Esses fluxos atravessam serviços (payments → canteen, canteen → notifications, canteen → reports) e precisam sobreviver a quedas de quem consome. É preciso um broker que entregue eventos de forma durável entre os serviços.

## Fatores de decisão

- Ordem por aluno: eventos da mesma carteira (ex.: recarga e compra) processados na sequência em que ocorreram.
- Vários consumidores independentes do mesmo evento, cada um com seu progresso (notifications e reports leem os mesmos eventos de compra).
- Retenção e replay: reconstruir projeções (reports) relendo eventos antigos.
- Entrega pelo menos uma vez, com consumidores idempotentes.
- Peso aceitável no ambiente local e no CI.

## Opções consideradas

- Kafka
- NATS JetStream

## Decisão

Escolhida: **Kafka**, com o cliente **franz-go**.

- Chave de partição: ID do aluno, garantindo ordem por carteira.
- Um consumer group por serviço consumidor.
- Consumidores idempotentes: cada evento carrega um ID único, e o efeito é aplicado no máximo uma vez (ver [0004](0004-outbox.md) e a chave de idempotência do ledger).
- Imagem local (`apache/kafka-native` ou Redpanda) decidida ao montar o compose.

## Consequências

- Positivas: ordem por chave, consumer groups e retenção configurável atendem diretamente os fatores de decisão; replay permite reconstruir projeções.
- Negativas: mais conceitos para operar (partições, offsets, rebalance); o broker não deduplica, então a idempotência fica inteira nos consumidores; container mais pesado que NATS.

## Prós e contras das opções

### Kafka

- Prós: log particionado com ordem por chave; consumer groups com offsets independentes; retenção longa e replay; ecossistema maduro (Kafka Connect, Schema Registry, Debezium).
- Contras: operação mais complexa; sem deduplicação no consumo; mais memória.

### NATS JetStream

- Prós: binário único e leve; janela de deduplicação no servidor por `Msg-Id`; configuração simples; escrito em Go.
- Contras: ordem e paralelismo por subject exigem mais desenho para chegar à ordem por aluno com vários consumidores; ecossistema menor para replay e integração (ex.: CDC).
