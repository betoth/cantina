---
name: adr
description: Registra uma decisão de arquitetura em docs/adr no formato MADR, em par com o usuário. Usar quando uma decisão técnica significativa for tomada ou precisar ser discutida (escolha de tecnologia, padrão, fronteira entre módulos, estratégia de consistência).
argument-hint: <título da decisão>
---

Registrar a decisão: **$ARGUMENTS**

ADRs são trabalho pareado: você propõe, o usuário revisa e decide.

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas. Nada de blocos de texto grandes: no chat, mostrar só o trecho em discussão; o texto completo fica no arquivo.

## Preparação

1. Se `$ARGUMENTS` estiver vazio, usar a decisão em discussão na conversa; se não estiver clara, perguntar qual e parar.
2. Ler o template [template.md](template.md), as ADRs existentes, `docs/domain.md`, `docs/conventions.md` e o discovery do tema em `docs/discovery/`, se existir: as opções e a comparação já feitas partem dele.
3. Avaliar se a decisão é significativa: difícil de reverter, com trade-off real, afeta o desenho. Se não for, propor uma linha na tabela Stack dos READMEs em vez de ADR e esperar o usuário decidir.
4. Se a decisão contradiz uma ADR aceita, ela substitui a anterior: avisar o usuário antes de seguir.

## Conteúdo

5. Começar pelo problema: por que essa decisão é necessária neste projeto, antes de qualquer ferramenta ou padrão.
6. Fatores de decisão: critérios técnicos, do domínio e de mercado (adoção, maturidade do ecossistema, suporte da comunidade).
7. Opções consideradas: só as que foram de fato discutidas. Se faltar uma alternativa óbvia que um revisor perguntaria, sugerir incluí-la e perguntar.
8. Consequências: positivas e negativas, ambas obrigatórias.

## Escrita

9. Próximo número: maior `NNNN` em `docs/adr/` + 1 (a primeira é 0001). Arquivo `NNNN-titulo-em-kebab-case.md`, em português. Status `proposta`, data de hoje.
10. Ao substituir uma ADR: na antiga, mudar só o status para `substituída por [NNNN](NNNN-titulo.md)`; não editar o restante.
11. Se a decisão adiciona, remove ou muda a comunicação entre containers, atualizar o diagrama de `docs/architecture.md`.
12. Se a decisão envolve ferramenta listada na tabela Stack, atualizar `README.md` e `README.pt-BR.md` com o link para a ADR.

## Fechamento

13. Pedir revisão. Status passa a `aceita` só quando o usuário aprovar.
