SHARED_FSPATH=./../../shared
BOILERPLATE_FSPATH=$(SHARED_FSPATH)/boilerplate

include $(BOILERPLATE_FSPATH)/help.mk
include $(BOILERPLATE_FSPATH)/os.mk
include $(BOILERPLATE_FSPATH)/gitr.mk
include $(BOILERPLATE_FSPATH)/tool.mk
include $(BOILERPLATE_FSPATH)/flu.mk
include $(BOILERPLATE_FSPATH)/go.mk

override FLU_SAMPLE_NAME =example
override FLU_LIB_NAME =client
override FLU_SAMPLE_FSPATH = $(PWD)/example/client

EXAMPLE_BIN_DIR=./bin-all
SERVER_BIN=$(EXAMPLE_BIN_DIR)/server
CLI_BIN=$(EXAMPLE_BIN_DIR)/cli

SERVER_PORT=8888
SERVER_ADDRESS=127.0.0.1:$(SERVER_PORT)
EXAMPLE_EMAIL = superadmin@getcouragenow.org
EXAMPLE_PASSWORD = superadmin

EXAMPLE_CFG_DIR = ./config
EXAMPLE_SYS_CORE_DB_ENCRYPT_KEY = mostSuperSecretKeyPlausible34@
EXAMPLE_SYS_CORE_CFG_PATH = $(EXAMPLE_CFG_DIR)/syscore.yml
EXAMPLE_SYS_ACCOUNT_CFG_PATH = $(EXAMPLE_CFG_DIR)/sysaccount.yml

EXAMPLE_SRC_DIR = ./example
SERVER_DIR = $(EXAMPLE_SRC_DIR)/server/go
CLI_DIR =  $(EXAMPLE_SRC_DIR)/cli/go

this-all: this-print this-gen this-build this-print-end
	
## Print all settings
this-print: 
	@echo
	@echo "-- SYS:SYS-ACCOUNT: start --"
	@echo SERVER_BIN: $(SERVER_BIN)
	@echo CLI_BIN: $(CLI_BIN)
	@echo

this-print-end:
	@echo
	@echo "-- SYS:SYS-ACCOUNT: end --"
	@echo
	@echo
	

## Builds everything
this-build: this-build-clean
	@echo "Building example golang binaries (CLI and Server)"
	#mkdir -p $(SERVER_BIN)
	#mkdir -p $(CLI_BIN)
	go build -v -o $(SERVER_BIN) $(SERVER_DIR)/main.go
	go build -v -o $(CLI_BIN) $(CLI_DIR)/main.go

this-build-clean:
	rm -rf $(SERVER_BIN) $(CLI_BIN)

this-gen:this-gen-clean this-config-gen
	# Go
	@go generate

this-gen-clean:
	# none

this-config-gen: this-config-clean this-config-dep
	@echo Generating Config
	@mkdir -p $(EXAMPLE_CFG_DIR)
	jsonnet -S $(SERVER_DIR)/sysaccount.jsonnet > $(EXAMPLE_SYS_ACCOUNT_CFG_PATH)
	jsonnet -S $(SERVER_DIR)/syscore.jsonnet -V SYS_CORE_DB_ENCRYPT_KEY=$(EXAMPLE_SYS_CORE_DB_ENCRYPT_KEY) > $(EXAMPLE_SYS_CORE_CFG_PATH)

this-config-clean:
	@echo Deleting previously generated config
	rm -rf $(EXAMPLE_SYS_CORE_CFG_PATH)
	rm -rf $(EXAMPLE_SYS_ACCOUNT_CFG_PATH)

this-config-dep:
	cd $(SERVER_DIR) && jb install && jb update

## Build and Run Server Example (Temporary until structure/stack is fixed)
this-ex-server-run:
	@echo Running Example Server
	$(SERVER_BIN) -c $(EXAMPLE_SYS_CORE_CFG_PATH) -a $(EXAMPLE_SYS_ACCOUNT_CFG_PATH)

this-ex-auth-signup:
	@echo Running Example Register Client
	$(CLI_BIN) sys-account auth-service register --email $(EXAMPLE_EMAIL) --password $(EXAMPLE_PASSWORD) --password-confirm $(EXAMPLE_PASSWORD) --server-addr $(SERVER_ADDRESS)

this-ex-auth-signin:
	@echo Running Example Login Client
	$(CLI_BIN) sys-account auth-service login --email $(EXAMPLE_EMAIL) --password $(EXAMPLE_PASSWORD) --server-addr $(SERVER_ADDRESS)

this-ex-accounts-list:
	@echo Running Example Accounts List
	$(CLI_BIN) sys-account account-service list-accounts --server-addr $(SERVER_ADDRESS)