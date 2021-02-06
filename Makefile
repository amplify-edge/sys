
SHARED_FSPATH=./../shared
#SHARED_FSPATH=./

BOILERPLATE_FSPATH=$(SHARED_FSPATH)/boilerplate

include $(BOILERPLATE_FSPATH)/help.mk
include $(BOILERPLATE_FSPATH)/os.mk
include $(BOILERPLATE_FSPATH)/gitr.mk
include $(BOILERPLATE_FSPATH)/tool.mk
include $(BOILERPLATE_FSPATH)/flu.mk
include $(BOILERPLATE_FSPATH)/go.mk


# remove the "v" prefix
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)

override FLU_SAMPLE_NAME =client
override FLU_LIB_NAME =client


CI_DEP=github.com/amplify-edge/sys
CI_DEP_FORK=github.com/joe-getcouragenow/sys

BIN_FOLDER=./bin-all
SDK_BIN=$(BIN_FOLDER)/sdk-cli
SERVER_BIN=$(BIN_FOLDER)/sys-main

EXAMPLE_SERVER_PORT=8888
EXAMPLE_SERVER_ADDRESS=127.0.0.1:$(EXAMPLE_SERVER_PORT)
EXAMPLE_EMAIL = gutterbacon@protonmail.com
EXAMPLE_PASSWORD = test1235
EXAMPLE_SUPER_EMAIL = superadmin@getcouragenow.org
EXAMPLE_SUPER_PASSWORD = superadmin
EXAMPLE_NEW_SUPER_EMAIL = gutterbacon@getcouragenow.org
EXAMPLE_NEW_SUPER_PASSWORD = SmokeOnTheWater70s
EXAMPLE_SYS_CORE_DB_ENCRYPT_KEY   = yYz8Xjb4HBn4irGQpBWulURjQk2XmwES
EXAMPLE_SYS_CORE_SENDGRID_API_KEY = SOME_SENDGRID_API_KEY
EXAMPLE_SYS_FILE_DB_ENCRYPT_KEY   = A9bhbid5ODrKQVvd9MY17P5MZ
EXAMPLE_SYS_FILE_CFG_PATH = ./config/sysfile.yml
EXAMPLE_SYS_CORE_CFG_PATH = ./config/syscore.yml
EXAMPLE_SYS_ACCOUNT_CFG_PATH = ./config/sysaccount.yml
EXAMPLE_SERVER_DIR = ./example/server
EXAMPLE_SDK_DIR = ./example/sdk-cli
# Please override this
EXAMPLE_BACKUP_FILE = ./db/backup/gcn.db_20201023108589.bak
EXAMPLE_CERT_DIR = ./certs
EXAMPLE_CERT_SERVER_NAME ?= $(EXAMPLE_CERT_DIR)/local.pem
EXAMPLE_CERT_SERVER_KEY ?= $(EXAMPLE_CERT_DIR)/local.key.pem
EXAMPLE_CA_ROOT_NAME ?= $(EXAMPLE_CERT_DIR)/rootca.pem
MKCERT_CA_ROOT_DIR = $(shell mkcert -CAROOT | printf %q)

EXAMPLE_ACCOUNT_ID = ???
EXAMPLE_VERIFY_TOKEN = ???
EXAMPLE_ORG_ID  = ???
EXAMPLE_PROJECT_ID = ???

this-all: this-print this-dep this-build this-print-end

## Print all settings
this-print: 
	@echo
	@echo "-- SYS: start --"
	@echo BIN_FOLDER: $(BIN_FOLDER)
	@echo SDK_BIN: $(SDK_BIN)
	@echo SERVER_BIN: $(SERVER_BIN)
	@echo CA_ROOT_DIR: $(MKCERT_CA_ROOT_DIR)
	@echo EXAMPLE_ACCOUNT_ID: ???
	@echo EXAMPLE_VERIFY_TOKEN: ???
	@echo EXAMPLE_ORG_ID: ???
	@echo EXAMPLE_PROJECT_ID: ???
	@echo EXAMPLE_SYS_CORE_DB_ENCRYPT_KEY: $(EXAMPLE_SYS_CORE_DB_ENCRYPT_KEY)
	@echo EXAMPLE_SYS_CORE_SENDGRID_API_KEY: $(EXAMPLE_SYS_CORE_SENDGRID_API_KEY)
	@echo

this-print-end:
	@echo
	@echo "-- SYS: end --"
	@echo
	@echo


this-dep:
	cd $(SHARED_FSPATH) && $(MAKE) this-all

this-dev-dep: this-gen-cert-dep
	## TODO Add to boot and version it.
	# GO111MODULE="on" go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb
	# brew install jsonnet jq

### BUILD

this-prebuild:
	# so the go mod is updated

this-build: this-build-delete #this-config-gen
	mkdir -p ./bin-all
	mkdir -p ./bench
	cd sys-account && $(MAKE) this-all
	cd sys-core && $(MAKE) this-all


this-build-delete:
	rm -rf $(BIN_FOLDER)

this-config-gen: this-config-delete this-config-dep
	@echo Generating Config
	@mkdir -p ./config
	jsonnet -S $(EXAMPLE_SERVER_DIR)/sysaccount.jsonnet \
		-V SYS_ACCOUNT_DB_ENCRYPT_KEY=$(EXAMPLE_SYS_CORE_DB_ENCRYPT_KEY) \
		-V SYS_ACCOUNT_FILEDB_ENCRYPT_KEY=$(EXAMPLE_SYS_FILE_DB_ENCRYPT_KEY) \
        -V SYS_ACCOUNT_SENDGRID_API_KEY=$(shell echo ${SENDGRID_API_KEY}) > $(EXAMPLE_SYS_ACCOUNT_CFG_PATH)
#	jsonnet -S $(EXAMPLE_SERVER_DIR)/syscore.jsonnet \
#		-V SYS_CORE_DB_ENCRYPT_KEY=$(EXAMPLE_SYS_CORE_DB_ENCRYPT_KEY) \
#		-V SYS_CORE_SENDGRID_API_KEY=$(shell echo ${SENDGRID_API_KEY}) > $(EXAMPLE_SYS_CORE_CFG_PATH)
#	jsonnet -S $(EXAMPLE_SERVER_DIR)/sysfile.jsonnet \
#		-V SYS_FILE_DB_ENCRYPT_KEY=$(EXAMPLE_SYS_FILE_DB_ENCRYPT_KEY) > $(EXAMPLE_SYS_FILE_CFG_PATH)

this-config-dep:
	cd $(EXAMPLE_SERVER_DIR) && jb install && jb update
	cd sys-account/service/go && jb install && jb update
	cd sys-core/service/go && jb install && jb update

this-config-delete:
	@echo Deleting previously generated config
	rm -rf $(EXAMPLE_SYS_CORE_CFG_PATH)
	rm -rf $(EXAMPLE_SYS_ACCOUNT_CFG_PATH)

this-gen-cert: this-gen-cert-delete
	@mkdir -p $(EXAMPLE_CERT_DIR)
	/usr/bin/env bash -c ./scripts/certgen.sh

this-gen-cert-delete:
	#mkcert -uninstall
	rm -rf $(EXAMPLE_CERT_DIR)/*.{pem,key,csr,crt}

