# Diário

Meu registro da evolução do projeto: o que fiz, o que decidi (e por que mudei de ideia) e o que aprendi ou revisei. Decisões significativas têm ADR própria em [docs/adr](adr/); aqui fica o caminho até elas.

## 2026-10-05 · Fase 0: harness, escopo e roadmap

### Feito

Dia de fundação, sem código de domínio. Escrevi o domínio, os casos de uso, os requisitos não funcionais, o roadmap das Fases 0 a 7 e as quatro primeiras ADRs. Montei o harness de IA (`CLAUDE.md`, convenções e as skills que conduzem cada documento) e o [quadro no GitHub Projects](https://github.com/users/betoth/projects/1), que vai acompanhar a execução das fases.

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
