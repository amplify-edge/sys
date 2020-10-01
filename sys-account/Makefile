# git include
BOILERPLATE_FSPATH=./../boilerplate
include $(BOILERPLATE_FSPATH)/help.mk
include $(BOILERPLATE_FSPATH)/os.mk
include $(BOILERPLATE_FSPATH)/tool.mk
include $(BOILERPLATE_FSPATH)/flu.mk
include $(BOILERPLATE_FSPATH)/go.mk
include $(BOILERPLATE_FSPATH)/grpc.mk

override FLU_SAMPLE_NAME =example
override FLU_LIB_NAME =client
override FLU_SAMPLE_FSPATH = $(PWD)/main/client

this-all: this-build
	
## Print all settings
this-print: ## print
	@echo -- REPO -- 
	@echo REPOSITORY: $(REPOSITORY)
	@echo

	$(MAKE) os-print
	
	$(MAKE) flu-print

	$(MAKE) flu-gen-lang-print

	#$(MAKE) srv-print


## This builds everything
this-build:
	@echo "Bu Build & Flutter"

	# Flutter
	#@mkdir -p client/lib/api/v2

	# Go
	@go generate

## Build and Run Server Example (Temporary until structure/stack is fixed)
this-run-example: this-build
	@echo Running Example Server
	$(GO_LIB_FSPATH)/bin-all/authserver

this-run-example-ui:
	@echo Running grpc-ui
	@grpcui -plaintext localhost:8888
