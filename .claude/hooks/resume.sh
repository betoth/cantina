#!/usr/bin/env bash
# SessionStart hook: reports the work in progress so a fresh session (startup or
# /clear) can resume with just "continue". Emits JSON: additionalContext goes to
# Claude's context, systemMessage is shown to the user.
# Never fails the session: errors are swallowed and the script exits 0.

command -v jq >/dev/null || exit 0
branch=$(git branch --show-current 2>/dev/null) || exit 0

context="Retomada: branch \`${branch:-detached}\`."
message=""

issue=$(sed -nE 's#^[a-z]+/([0-9]+)-.*#\1#p' <<<"$branch")
if [[ -n "$issue" ]] && command -v gh >/dev/null; then
  if body=$(gh issue view "$issue" --json number,title,body --jq '"#\(.number) \(.title)\n\(.body)"' 2>/dev/null | tr -d '\r') && [[ -n "$body" ]]; then
    header=${body%%$'\n'*}
    steps=$(awk '/^## Passos/{p=1; next} /^## /{p=0} p && /^- \[ \]/' <<<"${body#*$'\n'}")
    context+=$'\n'"Issue ${header}"$'\n'"Passos abertos:"$'\n'"${steps}"
    next=$(head -n 1 <<<"$steps" | sed -E 's/^- \[ \] //')
    if [[ -n "$next" ]]; then
      message="Retomando ${header}. Próximo passo: ${next}. Digite \"continue\"."
    else
      message="Retomando ${header}. Sem passos abertos."
    fi
  fi
fi

status=$(git status --short 2>/dev/null)
if [[ -n "$status" ]]; then
  context+=$'\n'"git status:"$'\n'"$(head -n 20 <<<"$status")"
fi

jq -n --arg ctx "$context" --arg msg "$message" '
  {hookSpecificOutput: {hookEventName: "SessionStart", additionalContext: $ctx}}
  + (if $msg == "" then {} else {systemMessage: $msg} end)'
exit 0
