
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


CI_DEP=github.com/getcouragenow/sys
CI_DEP_FORK=github.com/joe-getcouragenow/sys

BIN_FOLDER=./bin-all
SDK_BIN=$(BIN_FOLDER)/sdk-cli
SERVER_BIN=$(BIN_FOLDER)/sys-main

EXAMPLE_SERVER_PORT=8888
EXAMPLE_SERVER_ADDRESS=127.0.0.1:$(EXAMPLE_SERVER_PORT)
EXAMPLE_EMAIL = test@getcouragenow.org
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
	GO111MODULE="on" go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb
	brew install jsonnet jq

### BUILD

this-prebuild:
	# so the go mod is updated
	go get -u github.com/getcouragenow/sys-share

this-build: this-build-delete #this-config-gen

	mkdir -p ./bin-all
	mkdir -p ./bench

#	cd sys-account && $(MAKE) this-all
#	cd sys-core && $(MAKE) this-all

	go build -gcflags="all=-N -l" -o $(SDK_BIN) $(EXAMPLE_SDK_DIR)/main.go
	go build -gcflags="all=-N -l" -o $(SERVER_BIN) $(EXAMPLE_SERVER_DIR)/main.go

this-build-delete:
	rm -rf $(BIN_FOLDER)

this-config-gen: this-config-delete this-config-dep
	@echo Generating Config
	@mkdir -p ./config
	jsonnet -S $(EXAMPLE_SERVER_DIR)/sysaccount.jsonnet > $(EXAMPLE_SYS_ACCOUNT_CFG_PATH)
	jsonnet -S $(EXAMPLE_SERVER_DIR)/syscore.jsonnet \
		-V SYS_CORE_DB_ENCRYPT_KEY=$(EXAMPLE_SYS_CORE_DB_ENCRYPT_KEY) \
		-V SYS_CORE_SENDGRID_API_KEY=$(EXAMPLE_SYS_CORE_SENDGRID_API_KEY) > $(EXAMPLE_SYS_CORE_CFG_PATH)
	jsonnet -S $(EXAMPLE_SERVER_DIR)/sysfile.jsonnet \
		-V SYS_FILE_DB_ENCRYPT_KEY=$(EXAMPLE_SYS_FILE_DB_ENCRYPT_KEY) > $(EXAMPLE_SYS_FILE_CFG_PATH)

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

this-gen-cert-dep:
	brew install mkcert nss

### RUN

this-ex-sdk-run:
	$(SDK_BIN)

this-ex-server-run:
	mkdir -p db
	$(SERVER_BIN) -p $(EXAMPLE_SERVER_PORT) -a $(EXAMPLE_SYS_ACCOUNT_CFG_PATH) -c $(EXAMPLE_SYS_CORE_CFG_PATH)

this-ex-sdk-auth-signup:
	@echo Running Example Register Client
	$(SDK_BIN) sys-account auth-service register --email $(EXAMPLE_EMAIL) --password $(EXAMPLE_PASSWORD) --password-confirm $(EXAMPLE_PASSWORD) --server-addr $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson

this-ex-sdk-auth-signin:
	@echo Running Example Login Client
	# export access token to the .token file
	#$(SDK_BIN) sys-account auth-service login --email $(EXAMPLE_EMAIL) --password $(EXAMPLE_PASSWORD) --server-addr $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson | jq -r .accessToken > .token
	$(SDK_BIN) sys-account auth-service login --email $(EXAMPLE_EMAIL) --password $(EXAMPLE_PASSWORD) --server-addr $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson

this-ex-sdk-auth-signin-super:
	@echo Running Example Login Client
	# export access token to the .token file
	$(SDK_BIN) sys-account auth-service login --email $(EXAMPLE_SUPER_EMAIL) --password $(EXAMPLE_SUPER_PASSWORD) --server-addr $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson | jq -r .accessToken > .token

this-ex-sdk-auth-verify:
	@echo Running Example Verify Client
	$(SDK_BIN) sys-account auth-service verify-account --account-id $(EXAMPLE_ACCOUNT_ID) --verify-token $(EXAMPLE_VERIFY_TOKEN) --server-addr $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson

this-ex-sdk-accounts-new:
	@echo Running Example New Account
	$(SDK_BIN) sys-account account-service new-account -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --email gutterbacon@example.com --password gutterbacon123 --verified --created-at-seconds $(shell date +%s) --jwt-access-token $(shell awk '1' ./.token | tr -d '\n')

this-ex-sdk-accounts-list:
	@echo Running Example Accounts List
	$(SDK_BIN) sys-account account-service list-accounts --server-addr $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --jwt-access-token $(shell awk '1' ./.token | tr -d '\n')

this-ex-sdk-accounts-get:
	@echo Running Example Accounts Get
	$(SDK_BIN) sys-account account-service get-account -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --id $(EXAMPLE_ACCOUNT_ID) --jwt-access-token $(shell awk '1' ./.token | tr -d '\n')

this-ex-sdk-accounts-update:
	@echo Running Example Accounts Update
	$(SDK_BIN) sys-account account-service update-account -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --id $(EXAMPLE_ACCOUNT_ID) --jwt-access-token $(shell awk '1' ./.token | tr -d '\n') --disabled

this-ex-sdk-accounts-assign-super:
	@echo Assigning Account to Superuser
	$(SDK_BIN) sys-account account-service assign-account-to-role --assigned-account-id $(EXAMPLE_ACCOUNT_ID) --role-all --role-role 4 -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --jwt-access-token $(shell awk '1' ./.token | tr -d '\n')

this-ex-sdk-org-new:
	@echo Running Example Create Org
	$(SDK_BIN) sys-account org-proj-service -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --jwt-access-token $(shell awk '1' ./.token | tr -d '\n') new-org --name "ORG 1" --logo-url "https://avatars3.githubusercontent.com/u/59567775?s=200&v=4"

this-ex-sdk-org-get:
	@echo Running Example Get Org
	$(SDK_BIN) sys-account org-proj-service -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --jwt-access-token $(shell awk '1' ./.token | tr -d '\n') get-org --id $(EXAMPLE_ORG_ID)

this-ex-sdk-org-list:
	@echo Running Example List Org
	$(SDK_BIN) sys-account org-proj-service -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --jwt-access-token $(shell awk '1' ./.token | tr -d '\n') list-org

this-ex-sdk-org-update:
	@echo Running Example Update Org
	$(SDK_BIN) sys-account org-proj-service -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --jwt-access-token $(shell awk '1' ./.token | tr -d '\n') update-org --id $(EXAMPLE_ORG_ID) --name "ORG 2" --contact "contact@getcouragenow.org"

this-ex-sdk-project-new:
	@echo Running Example Create New Project
	$(SDK_BIN) sys-account org-proj-service -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --jwt-access-token $(shell awk '1' ./.token | tr -d '\n') new-project --org-id $(EXAMPLE_ORG_ID) --name PROJECT1 --logo-url "https://upload.wikimedia.org/wikipedia/commons/thumb/0/05/Go_Logo_Blue.svg/1200px-Go_Logo_Blue.svg.png"

this-ex-sdk-project-list:
	@echo Running Example List Project
	$(SDK_BIN) sys-account org-proj-service -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --jwt-access-token $(shell awk '1' ./.token | tr -d '\n') list-project

this-ex-sdk-project-get:
	@echo Running Example Get Project
	$(SDK_BIN) sys-account org-proj-service -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson --jwt-access-token $(shell awk '1' ./.token | tr -d '\n') get-project --id $(EXAMPLE_PROJECT_ID)

this-ex-sdk-bench: this-ex-sdk-bench-start this-ex-sdk-bench-01 this-ex-sdk-bench-02
	@echo -- Example SDK Benchmark: End --

this-ex-sdk-bench-start: 
	@echo -- Example SDK Benchmark: Start --

	@echo Running Example SDK Benchmark, Run server first!
	
this-ex-sdk-bench-01:
	# Small
	@echo USERS: 10
	@echo DB CONNECTIONS: 1
	$(SDK_BIN) sys-bench -e -t $(EXAMPLE_CA_ROOT_NAME) -s $(EXAMPLE_SERVER_ADDRESS) -j "./bench/fake-register-data.json" -p "../sys-share/sys-account/proto/v2/sys_account_services.proto" -n "v2.sys_account.services.AuthService.Register" -r 10 -c 1


this-ex-sdk-bench-02:
	# Medium
	@echo USERS: 100
	@echo DB CONNECTIONS: 10
	$(SDK_BIN) sys-bench -e -t $(EXAMPLE_CA_ROOT_NAME) -s $(EXAMPLE_SERVER_ADDRESS) -j "./bench/fake-register-data.json" -p "../sys-share/sys-account/proto/v2/sys_account_services.proto" -n "v2.sys_account.services.AuthService.Register" -r 100 -c 10

this-ex-sdk-bench-03:
	# Medium
	@echo USERS: 1000
	@echo DB CONNECTIONS: 100
	$(SDK_BIN) sys-bench -e -t $(EXAMPLE_CA_ROOT_NAME) -s $(EXAMPLE_SERVER_ADDRESS) -j "./bench/fake-register-data.json" -p "../sys-share/sys-account/proto/v2/sys_account_services.proto" -n "v2.sys_account.services.AuthService.Register" -r 1000 -c 100

this-ex-sdk-backup:
	$(SDK_BIN) db-admin-service backup -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson

this-ex-sdk-list-backup:
	$(SDK_BIN) db-admin-service list-backup -s $(EXAMPLE_SERVER_ADDRESS) --tls --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson

this-ex-sdk-restore:
	$(SDK_BIN) db-admin-service restore --backup-file $(EXAMPLE_BACKUP_FILE) --tls -s $(EXAMPLE_SERVER_ADDRESS) --tls-ca-cert-file $(EXAMPLE_CA_ROOT_NAME) -o prettyjson
