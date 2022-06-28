# Go executable to use, i.e. `make GO=/usr/bin/go1.18`
# Defaults to first found in PATH
GO?=go

include Makefile.*

# ----------------------------------------------------------------------
# Dependencies
# ----------------------------------------------------------------------

# To download more dependencies.
bin/bindl:
	mkdir -p ${PWD}/bin
	curl --location https://bindl.dev/bootstrap.sh | OUTDIR=${PWD}/bin bash

# ----------------------------------------------------------------------
# Build
# ----------------------------------------------------------------------

bin/snippets:
	@${GO} build -o bin/snippets ./cmd/snippets

.PHONY: bin/snippets-dev
bin/snippets-dev: bin/goreleaser
	bin/goreleaser build \
		--output bin/snippets \
		--single-target \
		--snapshot \
		--rm-dist

.PHONY: clean
clean:
	@([[ -f "bin/snippets" ]] && rm bin/snippets) || echo

.PHONY: rebuild
rebuild:
	@$(MAKE) clean
	@$(MAKE) bin/snippets

RUN_FLAGS +=
.PHONY: run
run: rebuild
	@./bin/snippets --debug

# ----------------------------------------------------------------------
# Tests
# ----------------------------------------------------------------------

.PHONY: test/unit
test/unit:
	${GO} test -race -short -v ./...

.PHONY: test/integration
test/integration:
	${GO} test -race -run ".*[Ii]ntegration.*" -v ./...

.PHONY: test/functional
test/functional:
	PATH=${PWD}/bin:${PATH} ${MAKE} bin/snippets
	PATH=${PWD}/bin:${PATH} ${GO} test -race -run ".*[Ff]unctional.*" -v ./...

.PHONY: test/all
test/all:
	${GO} test -race -v ./...

# ----------------------------------------------------------------------
# Lint
# ----------------------------------------------------------------------

.PHONY: lint
lint: bin/golangci-lint
	bin/golangci-lint run

.PHONY: lint/fix
lint/fix: bin/golangci-lint
	bin/golangci-lint run --fix

.PHONY: lint/gh-actions
lint/gh-actions: bin/golangci-lint
	bin/golangci-lint run --out-format github-actions
