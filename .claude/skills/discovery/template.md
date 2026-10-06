# Discovery: tema

- Motivação: decisões que este discovery prepara e por que são caras de reverter.
- Histórico:
  - AAAA-MM-DD: criação (perguntas 1 a N).
  - AAAA-MM-DD: aprofundamento (o que e por quê).
- Referências pesquisadas: produtos, projetos e artigos. Lista completa em [Fontes](#fontes).

Cada pergunta traz:

- como o mercado resolve;
- as opções comparadas;
- a recomendação para o projeto.

Recomendação não é decisão. As decisões saem (ver [Saídas](#saídas)):

- nas ADRs;
- no `domain.md`;
- nas specs.

## Resumo das recomendações

| # | Pergunta | Recomendação |
|---|---|---|
| 1 | ... | ... |

## 1. Pergunta

**Pergunta.** O que precisa ser decidido, quanto custa errar e quanto custa mudar depois.

**Eliminatórios e critérios.**

- Eliminatórios: o que qualquer opção precisa cumprir (invariantes, requisitos não funcionais, ADRs aceitas).
- Critérios:
  - técnicos;
  - do domínio;
  - de mercado.

**Mercado.** Como as referências resolvem, com fonte. Fato medido separado de opinião.

**Opções.**

| Critério | Opção A | Opção B (mais simples) |
|---|---|---|
| ... | ... | ... |

- Opção A:
  - Prós: um por item.
  - Contras: um por item.
  - Limitações: o que ela não resolve, em que escala deixa de servir.
- Opção B: mesma estrutura.

**Custo-benefício.** Custo de cada opção contra o benefício na escala real do projeto:

- construir;
- operar;
- reverter.

**Riscos e limitações.** Pre-mortem da opção recomendada e o que ela não resolve.

**Recomendação:** opção e motivo, em uma ou duas frases.

**O que mudaria a recomendação:** escala, requisito ou premissa que faria outra opção vencer.

## Saídas

| Saída | O quê |
|---|---|
| ADR | ... |
| `domain.md` | ... |
| Spec | ... |

## Fontes

- Referência: [título](url)
- Exemplos e anti-padrões (não referência): [projeto](url)
