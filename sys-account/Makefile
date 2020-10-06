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

EXAMPLE_EMAIL = superadmin@getcouragenow.org
EXAMPLE_PASSWORD = superadmin
EXAMPLE_SERVER_ADDRESS = 127.0.0.1:8888

this-all: this-print this-build this-print-end
	
## Print all settings
this-print: 
	@echo
	@echo "-- sys: start --"
	@echo SDK_BIN: $(SDK_BIN)
	@echo

this-print-end:
	@echo
	@echo "-- sys: end --"
	@echo
	@echo
	


## This builds everything
this-build:
	@echo "Build & Flutter"

	# Flutter
	#@mkdir -p flu/lib/api/v2

	# Go
	@go generate

## Build and Run Server Example (Temporary until structure/stack is fixed)
this-ex-server-run:
	@echo Running Example Server
	cd ./bin-all/server && rm -rf getcouragenow.db && ./main

this-ex-client-auth:
	@echo Running Example Register Client
	cd ./bin-all/cli && ./main sys-account auth-service register --email $(EXAMPLE_EMAIL) --password $(EXAMPLE_PASSWORD) --password-confirm $(EXAMPLE_PASSWORD) --server-addr $(EXAMPLE_SERVER_ADDRESS)
	@echo Running Example Login Client
	cd ./bin-all/cli && ./main sys-account auth-service login --email $(EXAMPLE_EMAIL) --password $(EXAMPLE_PASSWORD) --server-addr $(EXAMPLE_SERVER_ADDRESS)
