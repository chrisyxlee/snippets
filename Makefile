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

# To download Go dependencies.
.PHONY: gomod
gomod:
	go mod tidy

# ----------------------------------------------------------------------
# Build
# ----------------------------------------------------------------------

bin/snippets: gomod
	@${GO} build -o bin/snippets ./app/cmd/

.PHONY: bin/snippets-dev
bin/snippets-dev: bin/goreleaser gomod
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

run-local:
	@go run ./app

# ----------------------------------------------------------------------
# Tests
# ----------------------------------------------------------------------

.PHONY: test/unit
test/unit: gomod
	${GO} test -race -short -v ./...

# .PHONY: test/integration
# test/integration: gomod
# 	${GO} test -race -run ".*[Ii]ntegration.*" -v ./...
#
# .PHONY: test/functional
# test/functional: gomod
# 	PATH=${PWD}/bin:${PATH} ${MAKE} bin/snippets
# 	PATH=${PWD}/bin:${PATH} ${GO} test -race -run ".*[Ff]unctional.*" -v ./...

.PHONY: test/all
test/all: gomod
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
