VERSION                 := v0.0.1
TARGET					:= accesslog-exporter
TARGET_PATH             := cmd/accesslog-exporter/main.go

REVISION                := $(shell git rev-parse --short HEAD 2> /dev/null || echo 'unknown')
BRANCH                  := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null || echo 'unknown')

REPO_PATH               := github.com/ozonru/accesslog-exporter
LDFLAGS                 += -X $(REPO_PATH)/exposer.Version=$(VERSION)
LDFLAGS                 += -X $(REPO_PATH)/exposer.Revision=$(REVISION)
LDFLAGS                 += -X $(REPO_PATH)/exposer.Branch=$(BRANCH)
GOFLAGS                 := -ldflags "$(LDFLAGS)"

GOPATH					:= $(lastword $(subst :, ,$(GOPATH)))
GOOS					?= linux
GOARCH					?= amd64
GO						:= go
GOLANG_CI_LINT_BIN  	:= $(GOPATH)/bin/golangci-lint
GOLANG_CI_LINT_VERSION	:= $(shell $(GOLANG_CI_LINT_BIN) --version 2>/dev/null)

fmt:
	@echo ">> applying fmt command"
	$(GO) fmt ./...

vet:
	@echo ">> applying vet command"
	$(GO) vet ./...

test:
	@echo ">> run tests"
	$(GO) test ./...

lint: golang_ci_lint_bin
	@echo ">> applying golangci-lint command"
	$(GOLANG_CI_LINT_BIN) run

build:
	@echo ">> building binary..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) -o $(TARGET) $(TARGET_PATH)

bench:
	@echo ">> run benchmarks"
	$(GO) test -v -test.bench BenchmarkParse -test.run=noneedruningtests ./...

golang_ci_lint_bin:
	@echo ">> checking that golangci-lint exists"
ifdef GOLANG_CI_LINT_VERSION
	@echo ">> ok"
else
	$(error "please, install golangci-lint")
endif
