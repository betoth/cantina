# Convenções

Regras que valem para qualquer pessoa que escreve código ou documentação no projeto.

## Idiomas

- Código, identificadores, comentários, branches, commits e PRs em inglês.
- Documentação em `docs/` em português.
- `README.md` (inglês) e `README.pt-BR.md` (português) sempre sincronizados.

## Go

- Erros: envolver com `%w`; comparar com `errors.Is`/`errors.As`; erros de domínio como valores exportados do pacote do domínio.
- Sem `panic` fora de `main`.
- `context.Context` como primeiro parâmetro em operações de I/O.

## Testes

- Unitários table-driven.
- Integração com testcontainers.

## Git

GitHub Flow: a `main` está sempre estável; cada issue tem uma branch curta que volta para a `main` por PR. Ambientes, quando existirem, são promovidos pelo pipeline com a mesma imagem, não por branch.

```mermaid
flowchart LR
    A[Issue] --> B[Branch a partir da main]
    B --> C[Commits]
    C --> R{Revisão sem bloqueante?}
    R -- não --> C
    R -- sim --> D["PR com Closes #N (refino ou harness da retro: Refs #N)"]
    D --> E{"CI verde e testes do card marcados (regra da /cards)?"}
    E -- não --> C
    E -- sim --> F[Squash na main]
    F --> G[Issue fechada, card em Feito, branch apagada]
```

- Commits no formato [Conventional Commits](https://www.conventionalcommits.org): `feat(ledger): add double-entry transaction`.
- Branch por issue: `<tipo>/<issue>-<slug>`, ex.: `feat/12-register-purchase`.
- PR com título no formato do commit. Corpo com as seções `## Summary`, `## Tests` (testes do card) e `## Review` (cada achado da revisão e o que foi feito: corrigido, descartado com motivo ou sugestão pendente com motivo), e `Closes #N`. PR de refino (branch `docs/<N>-refine-<slug>`) e PR de harness vindo da `/retro` de outra issue (branch `feat/<N>-harness-<slug>`) não fecham a issue: citam `Refs #N` no lugar do `Closes`, o `## Tests` do PR de refino diz que não há testes de card, e o do PR de harness lista os checks rodados.
- Merge por squash: um commit por PR na `main`.

## Dados

- Dados da aplicação (catálogo, alunos etc.) sempre persistidos.
- Dados de demonstração criados pela API de cadastro, nunca por insert direto no banco. Seed direto só como simplificação temporária registrada no roadmap.
- Mocks e dados em memória só em testes.

## Documentação

- Enumeração com mais de dois itens, ou com itens que têm detalhe próprio, vai em lista, com subitens aninhados. Nada de itens encadeados numa frase com vírgula ou ponto e vírgula. Exceção: exemplos curtos entre parênteses.
- Em documento existente fora do formato, o trecho alterado é ajustado.
- Documentos descrevem o estado atual, sem "antes era X": o que mudou e por quê fica no diário, e a versão antiga no git (permalink, quando precisar citar). Exceções, por serem registro: o diário e os ADRs.

## Processo

- Criar arquivo, pasta, dependência ou ferramenta só quando for usado. Nada de stubs, pastas vazias ou placeholders.
