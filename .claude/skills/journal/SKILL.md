---
name: journal
description: Registra a entrada do dia no diário pessoal do projeto em docs/journal.md (feito, decisões e mudanças de ideia, aprendizados e revisões, dúvidas abertas). Usar ao concluir uma tarefa ou ao encerrar uma sessão de trabalho.
argument-hint: <assunto da entrada, opcional>
effort: low
---

Registrar a entrada do diário. Assunto: **$ARGUMENTS**

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas. Nada de blocos de texto grandes: no chat, mostrar só o trecho em discussão; o texto completo fica no arquivo.

## Preparação

1. Ler só as duas últimas entradas de `docs/journal.md` (`grep -n '^## ' docs/journal.md | tail -n 2` dá a linha de início) para seguir o formato e saber onde parou.
2. Levantar o que mudou desde a última entrada: a conversa atual, `git status` e `git log` desde a data da última entrada.
3. Se `$ARGUMENTS` estiver vazio, deduzir o assunto da tarefa da conversa.

## Formato

Entradas em ordem cronológica, a mais nova no fim do arquivo:

```
## AAAA-MM-DD · Fase N: assunto

### Feito
### Decisões
### Aprendizados e revisões
### Dúvidas abertas
```

O diário é pessoal: mostra a evolução do projeto para quem se interessar, não é relatório. O detalhe da execução fica nas issues.

- Uma entrada por dia. Se já existir entrada de hoje, completar essa entrada e ajustar o assunto do título para cobrir o dia. Dia com trabalho em duas fases: `Fases N e M`.
- Primeira pessoa, na voz do usuário ("decidi", "achei que", "percebi").
- **Feito:** dois ou três parágrafos curtos ou linhas sobre o essencial, com link para as issues e para os PRs que já existirem (o da própria tarefa ainda não existe quando a entrada é escrita). Sem listar arquivo por arquivo.
- **Decisões:** o que foi decidido e o caminho até lá, principalmente quando houve mudança de ideia. Inclui decisões de processo. Decisão com ADR: uma linha com o link, sem repetir a justificativa.
- **Aprendizados e revisões:** são do usuário: o que aprendeu e o que já sabia e revisou na sessão. Propor itens a partir das perguntas e dúvidas que ele levantou na sessão, mostrar a proposta e pedir que confirme ou reescreva com as próprias palavras.
- **Dúvidas abertas:** só dúvidas da tarefa do dia. Questões de domínio ficam em `docs/domain.md`. Omitir a seção se não houver nenhuma.

## Fechamento

- Mostrar ao usuário só as seções novas ou alteradas e pedir revisão.
