# Download Booty
BOOTY_URL := https://raw.githubusercontent.com/amplify-edge/booty/master/scripts

ifeq ($(OS),Windows_NT)
	BOOTY_URL:=$(BOOTY_URL)/install.ps1
else
	BOOTY_URL:=$(BOOTY_URL)/install.sh
endif

SHELLCMD :=
ADD_PATH :=
ifeq ($(OS),Windows_NT)
	SHELLCMD:=powershell -NoLogo -Sta -NoProfile -NonInteractive -ExecutionPolicy Unrestricted -Command "Invoke-WebRequest -useb $(BOOTY_URL) | Invoke-Expression"
	ADD_PATH:=export PATH=$$PATH:"/C/booty" # workaround for github CI
else
	SHELLCMD:=curl -fsSL $(BOOTY_URL) | bash
	ADD_PATH:=echo $$PATH
endif

all: print dep build print-end

## Print all settings
print:
	@echo
	@echo "-- SYS: start --"
	@echo BIN_FOLDER: $(BIN_FOLDER)
	@echo SDK_BIN: $(SDK_BIN)
	@echo SERVER_BIN: $(SERVER_BIN)
	@echo

print-end:
	@echo
	@echo "-- SYS: end --"
	@echo
	@echo

dep:
	@echo $(BOOTY_URL)
	$(SHELLCMD)
	$(ADD_PATH)
	booty install-all
	booty extract includes

### BUILD
build: build-delete #config-gen
	cd sys-account && $(MAKE) all
	cd sys-core && $(MAKE) all


build-delete:
	rm -rf $(BIN_FOLDER)

config-gen: config-delete config-dep
	@echo Generating Config
	@mkdir -p ./config
	@booty jsonnet -S $(EXAMPLE_SERVER_DIR)/sysaccount.jsonnet \
		-V SYS_ACCOUNT_DB_ENCRYPT_KEY=$(EXAMPLE_SYS_CORE_DB_ENCRYPT_KEY) \
		-V SYS_ACCOUNT_FILEDB_ENCRYPT_KEY=$(EXAMPLE_SYS_FILE_DB_ENCRYPT_KEY) \
        -V SYS_ACCOUNT_SENDGRID_API_KEY=$(shell echo ${SENDGRID_API_KEY}) > $(EXAMPLE_SYS_ACCOUNT_CFG_PATH)

config-dep:
	cd $(EXAMPLE_SERVER_DIR) && booty jb install && booty jb update
	cd sys-account/service/go && booty jb install && booty jb update
	cd sys-core/service/go && booty jb install && booty jb update

config-delete:
	@echo Deleting previously generated config
	rm -rf $(EXAMPLE_SYS_CORE_CFG_PATH)
	rm -rf $(EXAMPLE_SYS_ACCOUNT_CFG_PATH)

gen-cert: gen-cert-delete
	@mkdir -p $(EXAMPLE_CERT_DIR)

gen-cert-delete:
	rm -rf $(EXAMPLE_CERT_DIR)/*.{pem,key,csr,crt}

