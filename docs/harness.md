# Harness de IA

Como o Claude Code é configurado neste projeto: onde ficam as regras e como uma regra nova entra.

## Princípio

Violação mecânica vira check automático; julgamento vira regra.

- Mecânica: dá para detectar sem interpretar (link quebrado, formato de commit, código gerado desatualizado). Vira teste, lint ou passo do CI, e sai do texto das skills.
- Julgamento: exige ler o contexto (escopo de uma issue, se uma decisão merece ADR). Vira regra numa skill ou no agent `reviewer`.

## Onde fica cada coisa

- `CLAUDE.md`: o que vale em toda sessão (descrição, ponteiros, arquitetura, modos, interação, Git). Curto, porque é relido a cada request.
- Skills (`.claude/skills/<nome>/SKILL.md`): fluxos reutilizáveis, carregados só quando usados.
  - `/refine`: refino de um item até a issue Ready, numa sessão.
  - `/task`: execução de uma issue, do branch ao PR, e o Pronto.
  - `/retro`: depois da revisão, transforma o que travou ou falhou na sessão em check, regra ou ponteiro.
  - As demais mantêm um documento de `docs/` ou o quadro.
- Agents (`.claude/agents/`): só quando o contexto isolado traz vantagem:
  - olhar sem o viés da conversa que produziu o trabalho;
  - restrição de ferramentas (ex.: só leitura);
  - paralelismo e volume.
- Hooks (`.claude/hooks/`): o que precisa acontecer sempre, sem depender do Claude lembrar (ex.: retomada da sessão).
- `docs/conventions.md`: regras para qualquer pessoa que escreve no projeto. Não é importado pelo `CLAUDE.md`: o `reviewer` o cobra na revisão, e a `/task` também o lê antes de implementar. Cobrar só na revisão poupa contexto de quem implementa, mas aqui as convenções são curtas e lê-las antes evita rodadas de correção.

## Regras

- Fluxo reutilizável vai em skill (`.claude/skills/`), não em command (`.claude/commands/`). Command só quando tiver vantagem clara sobre skill, registrando o motivo.
- Por padrão, skills podem ser acionadas automaticamente pela `description`. `disable-model-invocation: true` quando só o usuário deve decidir o momento (ex.: `/checkpoint`, `/task`, `/retro`).
- Os documentos informam; as regras de manutenção de cada documento e os templates ficam na skill que o mantém.
