SHARED_FSPATH=./../includes

include $(SHARED_FSPATH)/help.mk

BIN_FOLDER=./bin-all
EXAMPLE_SERVER_BIN=$(BIN_FOLDER)/server
EXAMPLE_CLI_BIN=$(BIN_FOLDER)/cli

all: print gen build print-end

## Print all settings
print:
	@echo
	@echo "-- SYS:SYS-CORE: start --"
	@echo EXAMPLE_SERVER_BIN: $(EXAMPLE_SERVER_BIN)
	@echo EXAMPLE_SDK_BIN: $(EXAMPLE_CLI_BIN)
	@echo

print-end:
	@echo
	@echo "-- SYS:SYS-CORE: end --"
	@echo
	@echo

## This builds everything
build: build-clean
	@echo "Building example golang binaries (CLI and Server)"
	mkdir -p $(EXAMPLE_SERVER_BIN)
	mkdir -p $(EXAMPLE_CLI_BIN)

	go build -v -o $(EXAMPLE_SERVER_BIN) ./example/server/go/main.go
	go build -v -o $(EXAMPLE_CLI_BIN) ./example/cli/go/main.go

build-clean:
	rm -rf $(BIN_FOLDER)

gen:gen-clean
	# Go
	@go generate

gen-clean:
	# NONE

## Build and Run Server Example (Temporary until structure/stack is fixed)
ex-server-run:
	@echo Running Example Server
	cd ./bin-all/server && ./main

ex-ui-run:
	@echo Running grpc-ui
	@grpcui -plaintext localhost:8888
