Título: `[ID ou Tipo] Texto da entrega`

- Com ID no checkbox do roadmap: `[UC-OPER-01] Registrar compra`, `[RNF-CARGA-03] 20 compras simultâneas na mesma carteira`.
- Sem ID, o tipo no lugar: `[ADR] ...`, `[Técnico] ...`, `[Harness] ...`.
- Sub-issue: o ID da issue mãe. Fora do roadmap: `[Técnico] ...`, `[Bug] ...`.

Metadados: milestone da fase; label de tipo (`caso-de-uso`, `rnf`, `técnico`, `adr`, `harness`, `bug`); label de modo (`manual`, `pareado`, `delegado`); `não-planejado` quando o item entrou com a fase já em andamento.

Corpo:

```markdown
## Contexto

Por que esta entrega existe e onde se encaixa na fase. Uma a três frases.

## Referências

- Roadmap: [Fase N](https://github.com/betoth/cantina/blob/main/docs/roadmap.md#fase-n-tema)
- Casos de uso / RNF: UC-ATOR-NN, RNF-GRUPO-NN
- Spec: `docs/specs/NNNN-nome.md` (quando existir)
- ADRs: (quando houver)

## Escopo

Só em issue sem spec (harness, técnico). Com spec, ela está nas Referências.

- Entra: ...
- Não entra: ...
- Critérios: ...

## Passos

- [ ] Spec
- [ ] Testes
- [ ] Implementação
- [ ] `make check` verde
- [ ] Diário

## Testes

- [ ] (automático) comportamento verificado: `TestNome`
- [ ] (manual) roteiro: o que executar e o que deve ser observado

## Pronto quando

- Todos os testes acima marcados, antes do merge.
- Atende o critério de pronto N da fase.

## Fora de escopo

- O que fica para outra issue ou fase.
```
