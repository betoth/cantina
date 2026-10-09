---
name: discovery
description: Pesquisa como o mercado resolve um problema e compara opções de forma imparcial, com custo-benefício, antes de uma decisão cara de reverter, registrando em docs/discovery/<tema>.md, em par com o usuário. Usar antes de ADR ou spec quando houver alternativas reais a avaliar, ao pesquisar referências (produtos, repositórios, artigos) ou para aprofundar um discovery existente.
argument-hint: <tema ou pergunta>
---

Discovery: **$ARGUMENTS**

O discovery prepara decisões; não decide. Trabalho pareado: você pesquisa e propõe, o usuário revisa e decide. A saída é:

- uma recomendação fundamentada por pergunta;
- a lista do que vira ADR, regra no `domain.md` ou spec.

## Forma de conduzir

- Uma pergunta por vez. Esperar a resposta antes da próxima.
- Mensagens curtas: no chat, só a pergunta em discussão; o texto completo fica no arquivo.
- Pesquisar em fontes reais (documentação oficial, código, artigos com medição) antes de afirmar como o mercado faz. Na dúvida, buscar.
- Pesquisa web delegada a um subagent (`general-purpose`), que devolve só os fatos com a fonte de cada um. A página inteira não entra no contexto principal.

## Preparação

Documentos de `docs/` são lidos por seção: primeiro o índice de títulos (`grep -n '^#' <arquivo>`), depois só as seções ligadas ao tema. O documento inteiro só quando a tarefa o percorre todo (ex.: checagem de cobertura).

1. Se `$ARGUMENTS` estiver vazio, usar o tema em discussão na conversa; se não estiver claro, perguntar qual e parar.
2. Ler:
   - `docs/domain.md`;
   - `docs/non-functional-requirements.md`;
   - `docs/conventions.md`;
   - as ADRs;
   - os casos de uso ligados ao tema.
3. Procurar em `docs/discovery/` um arquivo do mesmo tema. Existindo, aprofundar nele (pergunta existente ou nova); só criar outro arquivo para tema diferente.
4. Listar as perguntas do discovery: só as que levam a decisão cara de reverter ou com trade-off real. Mostrar a lista e esperar o ok.

## Análise de cada pergunta

1. **Pergunta explícita**, com o custo de errar e a reversibilidade (o que custa mudar depois).
2. **Eliminatórios e critérios antes das opções.** Definir antes de olhar as opções evita escolher o critério que favorece a opção preferida.
   - Eliminatórios: o que a solução precisa cumprir de qualquer forma (invariantes, requisitos não funcionais, ADRs aceitas).
   - Critérios:
     - técnicos;
     - do domínio;
     - de mercado (adoção, maturidade do ecossistema, suporte da comunidade).
3. **Mercado:** como as referências resolvem, com fonte em cada afirmação. Separar fato medido (benchmark com máquina e carga) de opinião ou marketing.
4. **Opções:** ao menos duas alternativas sérias, sempre incluindo a mais simples que atende aos eliminatórios. Cada opção descrita na sua melhor versão, como um defensor dela a apresentaria.
5. **Comparação:** tabela critério × opção e, para cada opção:
   - prós;
   - contras;
   - limitações (o que ela não resolve, em que escala ou cenário deixa de servir).
6. **Custo-benefício** na escala real do projeto: custo contra o benefício efetivo. Solução de escala maior que a necessária custa sem retorno. Custos a considerar:
   - construir;
   - operar;
   - reverter.
7. **Imparcialidade:** buscar ativamente evidência contra a opção que estiver na frente; não ancorar na primeira ideia nem na preferência de quem pergunta (incluindo a do usuário e a sua). Se a evidência for fraca ou faltar, dizer.
8. **Antes de recomendar:**
   - *Pre-mortem:* "um ano depois, a opção recomendada deu errado; por quê?". Riscos que aparecerem entram em Riscos e limitações.
   - *Sensibilidade:* o que teria que mudar (escala, requisito, premissa) para outra opção vencer.
9. **Recomendação** em uma ou duas frases, com o motivo. Recomendação não é decisão: a decisão sai na ADR, no `domain.md` ou na spec.

## Referências fracas

- Projeto de estudo, tutorial ou repositório pequeno entra como exemplo ou anti-padrão, não como referência de mercado. Anti-padrão encontrado é registrado na pergunta que ele ilustra, com o motivo da falha.
- Achado de pesquisa feita fora deste fluxo (repositório, artigo, produto analisado na conversa) vai para o discovery do tema, encaixado na pergunta relevante, com a fonte em Fontes.

## Escrita

- Arquivo: `docs/discovery/<tema-em-kebab-case>.md`, em português, a partir do [template](template.md) (criar a pasta se não existir). Data de hoje no Histórico; ao aprofundar, acrescentar uma linha com a data e o motivo.
- Resumo das recomendações atualizado a cada pergunta fechada.
- Recomendação revista depois de fechada: atualizar a pergunta com o novo argumento e a nova recomendação, sem apagar o que levou à mudança.

## Fechamento

- Preencher Saídas: cada recomendação apontando para onde vira decisão:
  - ADR nova ou existente;
  - regra ou questão no `domain.md`;
  - spec.
- Mostrar o resumo das recomendações e as saídas e pedir revisão. Com o ok, sugerir a primeira saída a executar (`/adr`, `/domain` ou `/spec`).
