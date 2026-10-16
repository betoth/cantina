---
name: refine
description: Refina um item numa sessão só (entrega do roadmap, caso de uso, tarefa técnica ou de harness, bug) por grilling em rodadas, em par com o usuário, e cadastra a issue já pronta para começar, com sub-issues de tarefa quando for caso de uso. Usar quando o usuário quiser preparar, refinar ou detalhar um trabalho antes de executá-lo.
argument-hint: <ID do caso de uso, entrega do roadmap ou descrição>
---

Refinar: **$ARGUMENTS**

Nada vai para o GitHub antes de estar refinado. O roadmap é o backlog; a issue nasce pronta para começar (Ready). Todo o refino roda nesta sessão, e o resultado fica no corpo da issue, não em arquivo.

## Forma de conduzir (grilling)

- Em rodadas. Cada rodada traz todas as decisões desbloqueadas de uma vez (as que não dependem de resposta pendente):
  - numeradas;
  - cada uma com contexto curto e a resposta recomendada;
  - "sim" aceita todas as recomendações; o usuário responde só as que muda (ex.: "3: não, X").
- Com a recomendação, cada decisão vira pergunta fechada; por isso a rodada agrupa todas, e a regra geral de perguntas abertas uma por vez não se aplica ao grilling. Discussão aberta que surgir sai da rodada e segue uma por vez.
- Fatos, o Claude busca (código, `docs/`, ADRs, `gh`, `/research`); só decisões vão ao dono. Não perguntar o que dá para descobrir. Busca longa vai para um subagent, e as decisões que não dependem dela seguem na rodada.
- A rodada seguinte traz só o que as respostas desbloquearam ou abriram. O grilling acaba quando nenhuma lacuna impede escrever o Ready e o dono confirma que o entendimento é o mesmo; só então a issue é escrita.
- Mensagens curtas: no chat, só as decisões da rodada; o texto completo vai na issue.

## Preparação

Documentos de `docs/` são lidos por seção: primeiro o índice de títulos (`grep -n '^#' <arquivo>`), depois só as seções ligadas ao tema.

1. Se `$ARGUMENTS` estiver vazio, usar o item em discussão na conversa; se não estiver claro, perguntar qual e parar.
2. Ler `docs/roadmap.md`, o [template de issue](../cards/template.md) e os documentos ligados ao item:
   - caso de uso: a linha em `docs/use-cases.md`, as seções do `docs/domain.md`, os requisitos de `docs/non-functional-requirements.md` da fase e das anteriores que se apliquem, as ADRs do tema;
   - harness ou técnico: os arquivos que o item muda.
3. Conferir se o item já tem issue (`gh issue list --search`). Se tiver, refinar e propor a edição do corpo em vez de criar outra; issue mãe com parte das sub-issues já criadas segue das que faltam.
4. Fora do roadmap (parte de uma entrega, item que atravessa várias, bug): seguir "Issue fora do roadmap" da skill `/cards`. Se for algo que se pediria como entrega do produto, propor primeiro a entrada no roadmap (`/roadmap`).

## Lacunas

5. Levantar para si as lacunas que impedem o Ready:
   - caso de uso:
     - pré-condições, fluxo principal, fluxos alternativos e exceções, resultado;
     - erros: o resultado esperado de cada um e a garantia de nenhum efeito colateral;
     - regras de negócio do `domain.md` e questões em aberto que afetam o caso;
     - requisitos não funcionais aplicáveis, inclusive os já entregues em fases anteriores: são verificados de novo aqui;
     - concorrência, quando houver estado compartilhado;
     - invariantes do ledger tocadas, citadas pelo número do `domain.md`;
   - caso de uso e técnico: onde testar (o seam), o mais alto possível e, de preferência, um só por caso (ex.: a porta do ledger, a API HTTP), preferindo os que já existem;
   - harness ou técnico: o que entra, o que não entra, como verificar, dependências;
   - bug: o teste que reproduz o defeito e falha hoje, antes de qualquer hipótese de causa;
   - todos: modo (manual, pareado, delegado, pela tabela do `CLAUDE.md`).
6. Fechar as lacunas pelo grilling.

## Durante o grilling

- Termo ou regra de domínio novo ou mudado, ou questão em aberto respondida: atualizar `docs/domain.md` na hora, pela `/domain`, e citar a seção na issue.
- Decisão que muda o texto de um item do roadmap (desta fase ou de outra): atualizar o item na hora, pela `/roadmap`.
- Edições em `docs/` durante o grilling ficam na árvore de trabalho, sem branch: a branch nasce no passo 10, com o número da issue mãe.
- Fato de mercado ou de ferramenta que uma decisão precisa: `/research`. O resultado entra na rodada como contexto da decisão.
- ADR (`/adr`) só quando a decisão é, ao mesmo tempo:
  - difícil de reverter;
  - surpreendente sem contexto;
  - resultado de um trade-off real.
- Fora disso, a decisão fica na seção Decisões da issue.

## Fatiamento (caso de uso)

- Uma issue mãe do caso de uso e uma sub-issue por tarefa, pelo [template](../cards/template.md).
- Cada tarefa é uma fatia vertical (tracer bullet): atravessa as camadas de que precisa e é verificável sozinha. Não fatiar por camada.
  - Responde "o que demonstro quando estiver pronta?" com um comportamento, não com uma camada.
  - Cabe numa sessão nova, que não viu o refino: a issue basta para executá-la.
  - Mudança que facilita as seguintes (prefactoring) vira a primeira sub-issue, quando houver.
- Critérios de aceite:
  - numerados `CA-01`, `CA-02`… em sequência única no caso de uso;
  - no formato dado/quando/então, cada um verificável por um teste;
  - cada um nomeia a observação que o mostraria falso e falha no commit de partida da sub-issue: critério já verdadeiro, ou que só outra sub-issue satisfaz, está errado;
  - cada requisito não funcional aplicável com ao menos um critério que o verifique;
  - todo critério coberto por uma sub-issue.
- Dependências entre sub-issues pelo "blocked by" nativo do GitHub.
- Antes do Fechamento, mostrar a lista de fatias (nome, o que demonstra, CAs, bloqueios) e ajustar granularidade e bloqueios com o dono.

## Ready

A issue está pronta quando tem:

- Caso de uso: objetivo, fluxos e decisões na mãe; critérios de aceite em cada sub-issue. Harness ou técnico: Escopo com o que entra, o que não entra e os critérios.
- Onde testar (tarefa, técnico): o seam combinado.
- Bug: o teste que reproduz o defeito, visto falhando.
- Testes concretos, cada um automático ou manual, com o que observar.
  - Automático por padrão, citando o `CA-NN` que cobre no nome ou comentário do teste.
  - Manual só com ganho sobre o automático: o comportamento depende de algo que o teste não alcança (render no GitHub, serviço externo real, automação do quadro) ou exige o olho do dono.
  - Eficientes:
    - cada teste cobre um critério que nenhum outro cobre;
    - o manual não repete o que um automático já verifica;
    - o roteiro manual é curto, com o comando exato e o que observar.
  - Manual que só se observa depois do merge: marcado `(manual, depois do merge)`.
- Pronto quando, citando o critério de pronto da fase que atende, se houver.
- Texto durável: comportamento e interfaces, sem caminho de arquivo nem número de linha, que mudam antes de a issue ser executada.
- Labels de tipo e de modo e milestone da fase.

## Fechamento

7. Escrever cada issue num arquivo do scratchpad (usado depois no `--body-file`), mostrar no chat, uma vez, o texto completo que foi escrito, sem reler o arquivo, e perguntar se pode cadastrar.
8. Com o ok, cadastrar pela skill `/cards`: milestone se faltar, issue mãe, sub-issues com o "blocked by", quadro, link no roadmap.
9. Em `docs/use-cases.md`, o ID do caso vira link para a issue mãe. Caso de uso com detalhe legado em `docs/use-cases/`, antes de apagar o arquivo:
   - questões em aberto dele vão para o `docs/domain.md`, pela `/domain`: respondidas, na seção do tema; abertas, em Questões em aberto;
   - nomes e IDs que ele definia (ex.: "spec 0003") são procurados em `docs/`; referência que fica (ex.: em ADR, que é registro) vai anotada no PR de refino;
   - o resto já está na issue mãe; apagar o arquivo.
10. Arquivos alterados no refino (`docs/domain.md`, ADR, detalhe legado apagado): branch `docs/<N>-refine-<slug>`, com `N` da issue mãe e slug em inglês (ex.: `docs/20-refine-register-purchase`); revisão pelo agent `reviewer` com as regras do passo Revisão da skill `/task`; PR de refino pela seção Git de `docs/conventions.md`, depois de perguntar.
11. Sugerir `/clear` e dar a frase de retomada: `/task #N` com a primeira issue a executar.
