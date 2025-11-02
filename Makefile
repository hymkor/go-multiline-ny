ifeq ($(OS),Windows_NT)
    SHELL=CMD.EXE
    SET=set
    NUL=NUL
    WHICH=where.exe
else
    SHELL=bash
    SET=export
    NUL=/dev/null
    WHICH=which
endif
ifndef GO
    SUPPORTGO=go1.20.14
    GO:=$(shell $(WHICH) $(SUPPORTGO) 2>$(NUL)|| echo go)
endif

VERSION:=$(shell git describe --tags 2>$(NUL) || echo v0.0.0)

all:
	$(GO) fmt ./...
	$(GO) build

test:
	$(GO) test -v ./...

demo:
	$(GO) run examples/example.go

release:
	pwsh tools/latest-notes.ps1 | gh release create -d --notes-file - -t $(VERSION) $(VERSION)


.PHONY: all try test demo
