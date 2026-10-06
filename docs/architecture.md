# Arquitetura

Visão geral do que roda no MVP e de como as peças se comunicam. Contextos e regras de negócio: [domain.md](domain.md). Decisões: [adr/](adr/). Quando cada peça entra: [roadmap.md](roadmap.md).

## Serviços

```mermaid
flowchart LR
    atores([Responsável, aluno,<br/>operador, admin])
    psp([PSP Pix])
    email([E-mail])

    subgraph cantina[Cantina]
        canteen[canteen<br/>núcleo + ledger + identity]
        payments[payments<br/>integração Pix]
        notifications[notifications<br/>envio de mensagens]
        kafka{{Kafka}}
    end

    atores -- HTTP --> canteen
    atores -- HTTP: caixa de avisos --> notifications
    psp <--> payments

    canteen -- fatos e pedidos de envio --> kafka
    payments -- cobrança criada,<br/>recarga confirmada --> kafka
    kafka -- recarga solicitada --> payments
    kafka -- cobrança criada,<br/>recarga confirmada --> canteen
    kafka -- pedidos de envio --> notifications
    notifications --> email
```

- Entre atores e serviços, HTTP; os atores falam com o canteen (e com o notifications só para a caixa de avisos). O payments só fala com o PSP e com o Kafka. Entre serviços, só Kafka.
- Entre contextos de domínio circulam fatos. Ao notifications vão pedidos de envio genéricos; ele não conhece o domínio.
- Cada serviço tem o próprio schema no PostgreSQL. Ledger e identity são módulos do canteen, cada um com schema próprio.

## Publicação de eventos

Vale para todo serviço que publica no Kafka.

```mermaid
flowchart LR
    servico[serviço] -- estado + evento<br/>na mesma transação --> pg[(PostgreSQL<br/>com outbox)]
    relay[relay] -- lê pendentes --> pg
    relay -- publica --> kafka{{Kafka}}
```

O trabalho do serviço termina no commit. O relay publica depois; se cair no meio, publica de novo, e os consumidores são idempotentes.

## Observabilidade

```mermaid
flowchart LR
    servicos[serviços] -- OTLP --> otel[OpenTelemetry Collector]
    otel --> jaeger[Jaeger: traces]
    otel --> prom[Prometheus: métricas]
```

A aplicação conhece só o OpenTelemetry; os backends são trocados na configuração do Collector.
