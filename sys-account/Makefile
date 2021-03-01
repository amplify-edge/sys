SHARED_FSPATH=./../includes

include $(SHARED_FSPATH)/help.mk

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
EXAMPLE_SYS_CORE_SENDGRID_API_KEY = SOME_SENDGRID_API_KEY
EXAMPLE_SYS_FILE_DB_ENCRYPT_KEY   = A9bhbid5ODrKQVvd9MY17P5MZ

EXAMPLE_SRC_DIR = ./example
SERVER_DIR = $(EXAMPLE_SRC_DIR)/server/go
CLI_DIR =  $(EXAMPLE_SRC_DIR)/cli/go

all: print gen build print-end
	
## Print all settings
print:
	@echo
	@echo "-- SYS:SYS-ACCOUNT: start --"
	@echo SERVER_BIN: $(SERVER_BIN)
	@echo CLI_BIN: $(CLI_BIN)
	@echo

print-end:
	@echo
	@echo "-- SYS:SYS-ACCOUNT: end --"
	@echo
	@echo
	

## Builds everything
build: build-clean
	@echo "Building example golang binaries (CLI and Server)"
	#mkdir -p $(SERVER_BIN)
	#mkdir -p $(CLI_BIN)
	go build -v -o $(SERVER_BIN) $(SERVER_DIR)/main.go
	go build -v -o $(CLI_BIN) $(CLI_DIR)/main.go

build-clean:
	rm -rf $(SERVER_BIN) $(CLI_BIN)

gen:gen-clean config-gen
	# Go
	@go generate

gen-clean:
	# none

config-gen: config-clean config-dep
	@echo Generating Config
	@mkdir -p $(EXAMPLE_CFG_DIR)
#	jsonnet -S $(SERVER_DIR)/sysaccount.jsonnet > $(EXAMPLE_SYS_ACCOUNT_CFG_PATH)
#	jsonnet -S $(EXAMPLE_SERVER_DIR)/syscore.jsonnet \
#		-V SYS_CORE_DB_ENCRYPT_KEY=$(EXAMPLE_SYS_CORE_DB_ENCRYPT_KEY) \
#		-V SYS_CORE_SENDGRID_API_KEY=$(shell echo ${SENDGRID_API_KEY}) > $(EXAMPLE_SYS_CORE_CFG_PATH)
#	jsonnet -S $(EXAMPLE_SERVER_DIR)/sysfile.jsonnet \
#		-V SYS_FILE_DB_ENCRYPT_KEY=$(EXAMPLE_SYS_FILE_DB_ENCRYPT_KEY) > $(EXAMPLE_SYS_FILE_CFG_PATH)

config-clean:
	@echo Deleting previously generated config
#	rm -rf $(EXAMPLE_SYS_CORE_CFG_PATH)
#	rm -rf $(EXAMPLE_SYS_ACCOUNT_CFG_PATH)

config-dep:
	cd $(SERVER_DIR) && booty jb install && booty jb update

## Build and Run Server Example (Temporary until structure/stack is fixed)
ex-server-run:
	@echo Running Example Server
	$(SERVER_BIN) -c $(EXAMPLE_SYS_CORE_CFG_PATH) -a $(EXAMPLE_SYS_ACCOUNT_CFG_PATH)

ex-auth-signup:
	@echo Running Example Register Client
	$(CLI_BIN) sys-account auth-service register --email $(EXAMPLE_EMAIL) --password $(EXAMPLE_PASSWORD) --password-confirm $(EXAMPLE_PASSWORD) --server-addr $(SERVER_ADDRESS)

ex-auth-signin:
	@echo Running Example Login Client
	$(CLI_BIN) sys-account auth-service login --email $(EXAMPLE_EMAIL) --password $(EXAMPLE_PASSWORD) --server-addr $(SERVER_ADDRESS)

ex-accounts-list:
	@echo Running Example Accounts List
	$(CLI_BIN) sys-account account-service list-accounts --server-addr $(SERVER_ADDRESS)