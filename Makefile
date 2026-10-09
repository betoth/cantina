.PHONY: tokencost costs

# Installs the tokencost CLI used by the Claude Code hooks.
tokencost:
	go -C tools install ./tokencost

# Updates docs/ai-costs.csv from the local Claude Code transcripts.
costs:
	go -C tools run ./tokencost collect -root "$(CURDIR)"
