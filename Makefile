.PHONY: test phase1-verify phase2-verify smoke-weaviate

GO ?= go

 test:
	$(GO) test ./...

phase1-verify:
	$(GO) test ./internal/api ./internal/extractors ./internal/engine

phase2-verify:
	$(GO) test ./internal/api ./internal/extractors ./internal/engine ./internal/repo

 smoke-weaviate:
	./scripts/smoke_weaviate.sh
