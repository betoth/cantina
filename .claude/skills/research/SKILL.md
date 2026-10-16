---
name: research
description: Pesquisa leve em fontes primárias (documentação oficial, código, artigos com medição) para responder a um fato que uma decisão precisa, com a fonte em cada afirmação, sem gravar arquivo. Usar durante o grilling da /refine, antes de uma ADR, ou quando o usuário pedir referências de como o mercado resolve um problema.
argument-hint: <pergunta>
---

Pesquisar: **$ARGUMENTS**

A pesquisa traz fatos; não decide. O resultado volta para quem pediu (rodada do grilling, ADR, conversa) e não vira arquivo.

## Forma de conduzir

- Se `$ARGUMENTS` estiver vazio, usar a pergunta em discussão na conversa; se não estiver clara, perguntar qual e parar.
- Uma pergunta concreta por pesquisa, ligada a uma decisão (ex.: "como o Stripe garante idempotência na criação de cobrança?").
- Pesquisa web delegada a um subagent (`general-purpose`), que devolve só os fatos com a fonte de cada um. A página inteira não entra no contexto principal.

## Fontes

- Primárias: documentação oficial, código do projeto citado, artigo com medição (benchmark com máquina e carga).
- Fonte em cada afirmação, com link.
- Separar fato medido de opinião ou marketing.
- Projeto de estudo, tutorial ou repositório pequeno entra como exemplo ou anti-padrão, não como referência de mercado.
- Evidência fraca ou ausente: dizer, em vez de completar com suposição.

## Resultado

- Resposta curta à pergunta, em lista, com a fonte em cada item.
- Opções encontradas, quando houver: ao menos duas, sempre incluindo a mais simples que atende, cada uma na sua melhor versão, com o que ela não resolve.
- Para onde vai:
  - grilling da `/refine`: contexto da decisão na rodada, e o essencial na seção Decisões da issue;
  - ADR: seção "Opções consideradas", com as fontes.
