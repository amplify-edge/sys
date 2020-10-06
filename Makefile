
SHARED_FSPATH=./../shared
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


SDK_BIN=$(PWD)/bin-all/sdk-cli
EXAMPLE_EMAIL = superadmin@getcouragenow.org
EXAMPLE_PASSWORD = superadmin
EXAMPLE_SERVER_ADDRESS = 127.0.0.1:8888

this-all: this-print this-dep this-build this-print-end

## Print all settings
this-print: 
	@echo
	@echo "-- SYS: start --"
	@echo SDK_BIN: $(SDK_BIN)
	@echo

this-print-end:
	@echo
	@echo "-- SYS: end --"
	@echo
	@echo


this-dep:
	cd $(SHARED_FSPATH) && $(MAKE) this-all

this-prebuild:
	# so the go mod is updated
	go get -u github.com/getcouragenow/sys-share

this-build:

	mkdir -p ./bin-all

	cd sys-account && $(MAKE) this-all
	cd sys-core && $(MAKE) this-all

	cd main/sdk-cli && go build -o $(SDK_BIN) .

this-sdk-run:
	$(SDK_BIN)

this-ex-server-run:
	cd sys-account && $(MAKE) this-ex-server-run

this-ex-client-auth:
	cd sys-account && $(MAKE) this-ex-client-auth

this-example-sdk-auth:
	@echo Running Example Register Client
	cd ./bin-all/ && ./sdk-cli sys-account auth-service register --email $(EXAMPLE_EMAIL) --password $(EXAMPLE_PASSWORD) --password-confirm $(EXAMPLE_PASSWORD) --server-addr $(EXAMPLE_SERVER_ADDRESS)
	@echo Running Example Login Client
	cd ./bin-all/ && ./sdk-cli sys-account auth-service login --email $(EXAMPLE_EMAIL) --password $(EXAMPLE_PASSWORD) --server-addr $(EXAMPLE_SERVER_ADDRESS)
