
.PHONY: help
help: ## Help message
	@awk 'BEGIN{FS=":.*[#]#"}/:.*[#]#/{printf("%s # %s\n", $$1, $$2)}' $(MAKEFILE_LIST) | column -t -s#


.PHONY: docs
docs: notes ## Generate docs

.PHONY: notes
notes: ## Update notes
	plantuml -tsvg -oimg docs/notes/*
