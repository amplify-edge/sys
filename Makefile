
# This make file uses composition to keep things KISS and easy.
# In the boilerpalte make files dont do any includes, because you will create multi permutations of possibilities.

BOILERPLATE_FSPATH=./boilerplate

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

this-all: this-print this-dep this-build this-print-end

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


## Get dependencies for building
this-dep: grpc-all-git-delete grpc-all-git-clone grpc-go-build grpc-grpcui-build grpc-protoc-build

	@echo Need to get tools 

	# Need shared repo build tools, NOT bs, because BS will becme part of SDK
	# Want to use make and not github actions
	# - Need to check if on a fork or not.
	# 1. Do a git clone to the corect GOPATH.
	# 2. check if the folder is there, and if not do a git clone.

	# So need to convert GRPC tools to install using hashicorp go getter
	# https://github.com/hashicorp/go-getter
	# do it as part of shared/tools
	# SDK can then build  


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

this-ex-ui-run:
	cd sys-account && $(MAKE) this-ex-ui-run