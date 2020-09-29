# https://github.com/mono0926/flutter_playground


# git include
BOILERPLATE_FSPATH=./../../../../boilerplate
include $(BOILERPLATE_FSPATH)/help.mk
include $(BOILERPLATE_FSPATH)/os.mk
include $(BOILERPLATE_FSPATH)/tool.mk
include $(BOILERPLATE_FSPATH)/flu.mk
include $(BOILERPLATE_FSPATH)/go.mk

LIB=github.com/mono0926/flutter_playground
LIB_FSPATH=$(GOPATH)/src/$(LIB)

# flu build
override FLU_SAMPLE_FSPATH = $(LIB_FSPATH)

this-print:
	@echo
	@echo LIB: $(LIB)
	@echo LIB_FSPATH: $(LIB_FSPATH)
	@echo
	@echo

this-dep:
	git clone https://$(LIB) $(LIB_FSPATH)
this-dep-del:
	rm -rf $(LIB_FSPATH)

## this-flu-web-run
this-vscode-add:
	code --add --reuse-window $(LIB_FSPATH)

## this-flu-web-run
this-flu-web-run:
	$(MAKE) flu-web-run 

## this-flu-desk-init
this-flu-desk-init:
	$(MAKE) flu-desk-init 
	cd $(LIB_FSPATH) && hover plugins get


## this-flu-desk-run
this-flu-desk-run:
	$(MAKE) flu-desk-run 