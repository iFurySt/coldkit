PROJECT ?=
SLUG ?=

.PHONY: init new-history new-plan test build

init:
	@if [ -z "$(PROJECT)" ]; then echo "usage: make init PROJECT=my-project"; exit 1; fi
	./scripts/init-project.sh "$(PROJECT)"

new-history:
	@if [ -z "$(SLUG)" ]; then echo "usage: make new-history SLUG=my-change"; exit 1; fi
	./scripts/new-history.sh "$(SLUG)"

new-plan:
	@if [ -z "$(SLUG)" ]; then echo "usage: make new-plan SLUG=my-plan"; exit 1; fi
	./scripts/new-exec-plan.sh "$(SLUG)"

test:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/ck ./cmd/ck
	go build -o bin/ck-mcp ./cmd/ck-mcp
