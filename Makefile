REPO := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

BITCHAN_WASM := $(REPO)/webroot/static/bitchan.wasm
WASM_EXEC_JS := $(REPO)/webroot/static/wasm_exec.min.js
GOROOT = $(shell go env GOROOT)
EXE = bitchand

all: mistake

mistake: $(BITCHAN_WASM) $(EXE)

$(EXE):
	go build -o $(EXE)

$(BITCHAN_WASM):
	GOOS=js GOARCH=wasm go build -o '$(BITCHAN_WASM)' github.com/majestrate/bitchan/js
	cp '$(GOROOT)/misc/wasm/wasm_exec.js' '$(WASM_EXEC_JS)'

clean: repent

repent:
	rm -f '$(BITCHAN_WASM)' '$(WASM_EXEC_JS)' '$(EXE)'
	go clean -a
