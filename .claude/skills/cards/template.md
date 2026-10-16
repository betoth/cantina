Título: `[ID ou Tipo] Texto da entrega`

- Com ID no checkbox do roadmap: `[UC-OPER-01] Registrar compra`, `[RNF-CARGA-03] 20 compras simultâneas na mesma carteira`.
- Sem ID, o tipo no lugar: `[ADR] ...`, `[Técnico] ...`, `[Harness] ...`.
- Sub-issue: o ID da issue mãe. Fora do roadmap: `[Técnico] ...`, `[Bug] ...`.

Metadados:

- Milestone da fase.
- Label de tipo: `caso-de-uso`, `rnf`, `técnico`, `adr`, `harness` ou `bug`.
- Label de modo: `manual`, `pareado` ou `delegado`.
- `não-planejado` quando o item entrou depois do roadmap validado.
- Sub-issue: ligada à mãe e com as dependências pelo "blocked by" nativo do GitHub.

Corpo: seção com tipo entre parênteses no título só entra nas issues desse tipo, e o marcador sai do título ao cadastrar. Seção sem marcador entra em todas.

- Caso de uso: issue mãe.
- Tarefa: sub-issue de caso de uso.
- Harness ou técnico: issue única, também para RNF, ADR e bug.

```markdown
## Contexto

Por que esta entrega existe e onde se encaixa na fase. Uma a três frases.

## Referências

- Roadmap: [Fase N](https://github.com/betoth/cantina/blob/main/docs/roadmap.md#fase-n-tema)
- Casos de uso / RNF: UC-ATOR-NN, RNF-GRUPO-NN
- Issue mãe: #N (tarefa)
- Domínio: seções do `docs/domain.md` que se aplicam (caso de uso, tarefa)
- ADRs: (quando houver)

## Objetivo (caso de uso)

Como <ator>, quero <ação>, para que <benefício>.

## Fluxos (caso de uso)

Resumidos.

- Principal: ...
- Alternativos e exceções:
  - **<situação>:** o que acontece.

## Decisões (caso de uso)

Decisões do grilling que não viraram ADR, cada uma com o porquê.

- ...

## Tarefas (caso de uso)

As sub-issues, na ordem de execução, com os critérios que cobrem.

- [ ] #N Nome da fatia (CA-01, CA-02)

## Fatia (tarefa)

A fatia vertical (tracer bullet) que esta tarefa entrega, verificável sozinha. Uma a três frases.

- Demonstração: o comportamento que se mostra quando ela estiver pronta.
- Onde testar: o seam combinado no refino (ex.: porta do ledger, API HTTP).

## Critérios de aceite (tarefa)

Numerados na sequência do caso de uso; cada um vira ao menos um teste.

- **CA-01.** **Dado** ..., **quando** ..., **então** ...

## Escopo (harness ou técnico)

- Entra: ...
- Não entra: ...
- Critérios: ...
- Onde testar: o seam combinado no refino (técnico).

## Passos (tarefa, harness ou técnico)

- [ ] Branch e card em Em andamento
- [ ] Ciclo por critério (teste vermelho, código mínimo, repete)
- [ ] Verificação (`make check` verde)
- [ ] Checkpoint (perguntado ao dono)
- [ ] Diário e roadmap
- [ ] Revisão (agent `reviewer`)
- [ ] Retro (perguntada ao dono)
- [ ] PR

## Testes (tarefa, harness ou técnico)

- [ ] (automático) CA-01: comportamento verificado: `TestNome`
- [ ] (automático) bug: teste que reproduz o defeito, vermelho antes da correção: `TestNome`
- [ ] (manual) roteiro: o que executar e o que deve ser observado
- [ ] (manual, depois do merge) roteiro que só se observa depois do merge

## Pronto quando

- Caso de uso: todas as sub-issues fechadas.
- Tarefa, harness ou técnico:
  - todos os testes acima marcados antes do merge, exceto os `(manual, depois do merge)`, marcados no passo PR mergeado da skill `/cards`;
  - revisão sem bloqueante pendente.
- Atende o critério de pronto N da fase.

## Fora de escopo

- O que fica para outra issue ou fase.
```
