# Diário

Meu registro da evolução do projeto: o que fiz, o que decidi (e por que mudei de ideia) e o que aprendi ou revisei. Decisões significativas têm ADR própria em [docs/adr](adr/); aqui fica o caminho até elas.

## 2026-10-05 · Fases 0 e 1: harness, escopo, roadmap e reviewer

### Feito

Dia de fundação, sem código de domínio. Escrevi o domínio, os casos de uso, os requisitos não funcionais, o roadmap das Fases 0 a 7 e as quatro primeiras ADRs. Montei o harness de IA (`CLAUDE.md`, convenções e as skills que conduzem cada documento) e o [quadro no GitHub Projects](https://github.com/users/betoth/projects/1), que vai acompanhar a execução das fases.

Fechei a Fase 0 e abri a Fase 1 pelo harness: a skill `/refine` ([#3](https://github.com/betoth/cantina/issues/3), [PR #4](https://github.com/betoth/cantina/pull/4)), que prepara cada item antes de virar issue, e o agent revisor ([#2](https://github.com/betoth/cantina/issues/2)), testado numa branch descartável com problemas plantados.

### Decisões

- Stack: goose, sqlc + pgx, Kafka e outbox, cada um com sua ADR em [docs/adr](adr/). O resto (Go 1.27, golangci-lint, `tools/go.mod`) ficou numa linha da tabela de stack do README, porque ADR é só para decisão difícil de reverter.
- Comecei querendo deixar a estrutura do projeto pronta e voltei atrás: arquivo, pasta ou ferramenta só nasce quando for usado. Makefile, compose e CI ficam para a Fase 1.
- Pensei em commands para os fluxos do Claude, mas skill é acionada sozinha pela descrição e carrega o próprio template. Fiquei com skills. As regras de manutenção de cada documento moraram nos documentos por um tempo; levei para as skills, e os documentos só informam.
- Pensei em listar os casos de uso no domínio. Ficaram num arquivo próprio, com IDs, e o detalhe de cada um vai num arquivo separado quando for detalhado, para não ter que ler tudo de uma vez.
- Coloquei auditoria, logs e traces na Fase 1, junto com o ledger. Mas cada funcionalidade nova precisa provar que cumpre esses requisitos; não dá para assumir que funciona só porque a base existe.
- O notifications começou lendo os eventos do canteen. Separei a responsabilidade: ele só entrega o que recebe, e quem conhece as preferências do responsável é o canteen.
- O seed ia conviver com os cadastros. Decidi que some na Fase 5, para não esconder erro de cadastro.
- Diagrama: o de containers fica em `architecture.md`; o fluxo de cada funcionalidade fica na spec, desenhado na hora de implementar.
- Primeiro pensei em acompanhar a execução num checklist local. Troquei por issues no GitHub: milestone por fase, issue por entrega, título `[ID ou Tipo] Texto da entrega`, em português. O roadmap continua sendo o plano.
- Nem toda issue nasce de uma entrega do roadmap. Parte de uma entrega vira sub-issue; o que atravessa várias entregas (ou é bug) vira issue própria, ligada às entregas que atende.
- Ia testar o card depois do merge, como regressão. Separei: o aceite do card roda antes do merge, na branch, e a regressão roda na `main` ao fechar a fase.
- Git: Conventional Commits, branch por issue e squash no merge, tudo em inglês. Pensei numa skill para isso, mas é convenção para qualquer pessoa, então foi para `conventions.md`; skill só se abrir PR virar receita repetitiva.
- Faltava o nível entre o caso de uso e o código: no roadmap não existe "criar ledger", mas a compra precisa dele. A cadeia ficou entrega → caso de uso → specs → tarefas. O caso de uso lista as specs (o ledger é uma delas), e cada spec termina com as tarefas de implementação, que viram issues.
- Pensei em GitFlow com branches de DEV e HML, para estudo. Fiquei com GitHub Flow: branch por ambiente diverge com o tempo, e aqui ainda não existe ambiente. Quando houver, a promoção entre ambientes vai ser pelo pipeline, com a mesma imagem.
- Criei a primeira issue da Fase 1 (o agent reviewer) e percebi que ela nasceu sem escopo: "definir escopo" era o primeiro passo da execução. Pensei numa coluna Backlog para issues não refinadas, mas criar o card para depois editá-lo é retrabalho. Fiquei com refinar antes de criar: o roadmap é o backlog, e a skill `/refine` conduz o refinamento até a issue nascer pronta. Ela entrou na fase como não planejada, e marquei isso no roadmap e com uma label.
- O revisor não repete regra nenhuma: para cada artefato, lê a skill ou o documento que define as regras dele, e só traz de próprio a postura de revisor. Roda antes de todo PR; bloqueante tem que ser resolvido ou descartado com motivo, e sugestão pendente pede minha confirmação. No teste, achou os problemas plantados e ainda mostrou uma falha no fluxo, que punha o diário depois do PR; o diário passou para antes da revisão.
- Ajustei o fluxo por tarefa ao longo das revisões: o checkpoint vem antes do diário, para os pontos dele entrarem na mesma entrada; a marcação do roadmap deixou de ser depois do merge, porque sobrava para o PR seguinte, e passou para antes da revisão; o PR ganhou a seção `## Review`; e limitei a revisão a três rodadas com bloqueante antes de eu decidir, para não virar laço.
- Quis exigir aprovação de review no GitHub, mas o autor não pode aprovar o próprio PR; num projeto solo isso travaria todo merge. A revisão fica garantida pelo fluxo e registrada no PR.
- Por enquanto, toda alteração no GitHub feita pelo Claude passa por mim antes. Afrouxo quando ganhar confiança no fluxo.
- Só CI por enquanto. CD e ambientes de DEV e HML ficam para depois.

### Aprendizados e revisões

- Go não tem ferramenta nativa de migrations nem linter completo: só `gofmt` e `go vet`. golangci-lint agrega dezenas de linters e é o padrão.
- A diretiva `go` no `go.mod` faz qualquer Go ≥ 1.21 baixar o toolchain certo automaticamente.
- `go get -tool` adiciona as dependências da ferramenta ao `go.mod`; a CLI do goose trouxe drivers de vários bancos. Daí o módulo separado.
- ADR no formato MADR: contexto, fatores de decisão, opções, decisão, consequências. Começar pelo problema, não pela ferramenta.
- Dual write (gravar no banco e publicar no broker em sequência) perde ou inventa eventos; o outbox resolve gravando o evento na mesma transação.
- Kafka não deduplica no consumo; com outbox, a entrega é pelo menos uma vez, então os consumidores precisam ser idempotentes.
- Skill é acionada sozinha pela descrição e pode ter arquivos de apoio, como um template; command só roda quando chamado.
- O padrão de mercado para casos de uso é user story com critérios de aceite no Jira ou no Issues; spec no repositório faz sentido quando a IA precisa ler.
- Com o Kafka parado, a compra continua funcionando: é síncrona e só depende do banco. O evento espera no outbox.
- Agendador de jobs com várias instâncias precisa de lock, para cada job rodar uma vez por ciclo.
- C4 mostra a estrutura (containers); o fluxo do que é executado fica num fluxograma, dentro da spec.
- No GitHub Projects, automações e visões só são configuradas pela interface; o `gh` e a API cuidam de issues, campos e itens.
- Teste de aceite do card roda antes do merge, na branch; regressão roda depois, na `main`, ao fechar a fase.

## 2026-10-06 · Fase 1: modelo contábil e concorrência do ledger, skill /discovery

### Feito

Refinei o ledger antes da spec ([#6](https://github.com/betoth/cantina/issues/6)): detalhei o UC-OPER-01, fiz o discovery do ledger comparando ledgers de mercado e dois projetos de estudo, e registrei duas ADRs, o [modelo contábil](adr/0005-modelo-contabil-do-ledger.md) e a [concorrência no saldo](adr/0006-concorrencia-no-saldo.md). O `domain.md` foi alinhado às duas.

Criei a skill `/discovery` ([#7](https://github.com/betoth/cantina/issues/7)), a partir do formato que surgiu no discovery do ledger. Ela conduz a pesquisa de mercado e a comparação de opções antes de uma decisão cara de reverter, e entrou no fluxo por entrega entre o refinamento e as ADRs e specs.

Testei usando a skill para adaptar o discovery do ledger ao template e aprofundar concorrência ([#6](https://github.com/betoth/cantina/issues/6)).

### Decisões

- Pensei em fazer um discovery novo só de concorrência. Fiquei com aprofundar o do ledger: o assunto já estava lá, e dois arquivos sobre o mesmo tema espalhariam a análise.
- A skill ficou numa issue própria, separada do ledger, para cada PR tratar de uma coisa.
- Pedi imparcialidade e custo-benefício explícitos. A skill exige:
  - eliminatórios e critérios antes das opções;
  - ao menos duas alternativas sérias, incluindo a mais simples;
  - fonte em cada afirmação de mercado;
  - pre-mortem;
  - o que mudaria a recomendação.
- O revisor apontou que o template não batia com o discovery do ledger. Entre aceitar os dois formatos e adaptar o ledger, adaptei o ledger.
- Enumeração vai em lista, com subitens, nunca encadeada numa frase. Virou convenção para todo documento; no que já existe, o trecho alterado é ajustado.
- A marca de não planejado no roadmap valia só para item que entrava numa fase em andamento. Mudei para todo item novo depois que o roadmap foi validado, em qualquer fase, para separar o plano original do que surgiu na execução.

- Os dois projetos de estudo que analisei ensinaram mais pelos erros: lock no Redis liberado antes de gravar, idempotência por hash em cache de memória, ledger que se dizia imutável mas apagava lançamentos. Ficaram no discovery como anti-padrões.
- Hash encadeado com digest externo, como no SQL Server Ledger, ficou para a v2: protege contra quem tem acesso ao banco, que não é o risco de uma cantina, e o encadeamento serializaria as escritas do recreio.
- O ledger não tem API HTTP. Quem cria transação é o caso de uso (compra, estorno, recarga), dentro da transação do banco; ajuste manual, se existir, vira operação de negócio própria, com permissão e auditoria.
- No começo achei estranho a entrada de Pix ficar negativa. Fiquei com o lado normal na conta e direção com valor positivo no lançamento, como a contabilidade e o Modern Treasury ([ADR 0005](adr/0005-modelo-contabil-do-ledger.md)).
- Achei que as contas quentes seriam um problema. A estimativa mostrou que, no volume do projeto, a receita fica ocupada menos de 1% do tempo; fiquei com tudo materializado e síncrono, medindo na Fase 6 ([ADR 0006](adr/0006-concorrencia-no-saldo.md)).
- A revisão achou falhas reais na primeira versão da ADR 0006, e as corrigi antes do merge:
  - a recusa por saldo podia acontecer depois de a receita já ter sido creditada; resolvi com savepoint;
  - compra e estorno podiam travar um ao outro pelo limite diário; resolvi com uma ordem global de travamento;
  - retentativas e estornos simultâneos passavam por checagens de leitura; passaram a ser garantidos na escrita.

### Aprendizados e revisões

- Spike: trabalho com prazo cuja saída é aprendizado e uma recomendação, não código.
- Definir os critérios antes de olhar as opções evita escolher o critério que favorece a opção preferida.
- Pre-mortem: imaginar que a decisão deu errado em um ano e perguntar por quê. Faz aparecer riscos que a análise a favor esconde.
- Análise de sensibilidade: dizer o que teria que mudar para outra opção vencer.
- Lado normal: cada conta cresce num lado. O que a cantina tem (dinheiro no banco) cresce no débito; o que ela deve (carteira do aluno) e o que ganha (receita) crescem no crédito. Lançamento do mesmo lado soma; do lado oposto, subtrai.
- O "crédito" do extrato bancário é o ponto de vista do banco, para quem meu saldo é uma dívida. No livro da cantina, o dinheiro que entra no banco é débito.
- Transação é o fato (o motivo, a chave de idempotência); lançamento é o efeito em cada conta. Uma transação tem dois ou mais lançamentos.
- A versão da conta sobe a cada lançamento e mostra se algum se perdeu; com o saldo resultante gravado, o extrato não precisa recalcular nada.
- Dinheiro em centavos inteiros: `float` erra e Go não tem decimal nativo.
- Conta quente não vem da partida dobrada, vem do saldo materializado: inserir lançamento não disputa nada; atualizar a mesma linha de saldo, sim.
- Update condicional em `READ COMMITTED`: o segundo `UPDATE` espera o primeiro e reavalia o `WHERE` com o saldo novo; a recusa por saldo vira "0 linhas", não erro.
- Toda garantia sob concorrência tem que estar na escrita. Uma leitura antes da escrita deixa uma janela aberta.
- Deadlock se evita travando as linhas sempre na mesma ordem, em todos os fluxos.
