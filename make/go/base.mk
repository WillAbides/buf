# Managed by makego. DO NOT EDIT.

# Must be set
$(call _assert_var,MAKEGO)
# Must be set
$(call _assert_var,MAKEGO_REMOTE)
# Must be set
$(call _assert_var,PROJECT)
# Must be set
$(call _assert_var,GO_MODULE)

UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

ENV_DIR := .env
ENV_SH := $(ENV_DIR)/env.sh
ENV_BACKUP_DIR := $(HOME)/.config/$(PROJECT)/env

TMP := .tmp

# Settable
# These are exact file paths that are added to .gitignore and .dockerignore
FILE_IGNORES := $(FILE_IGNORES) $(ENV_DIR)/ $(TMP)/
# Settable
# These will be added without any modification to .gitignore
GIT_FILE_IGNORES ?=
# Settable
# These will be added without any modification to .dockerignore
DOCKER_FILE_IGNORES ?=
# Settable
CACHE_BASE ?= $(HOME)/.cache/$(PROJECT)

CACHE := $(CACHE_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
CACHE_BIN := $(CACHE)/bin
CACHE_INCLUDE := $(CACHE)/include
CACHE_VERSIONS := $(CACHE)/versions
CACHE_ENV := $(CACHE)/env
CACHE_GO := $(CACHE)/go

# CACHE_GOBIN is the location where binaries are installed for Golang projects
# This is as opposed to CACHE_BIN, where dependencies binaries are installed
# The separation is needed for i.e. buf, where we need to bootstrap with a
# download from releases, but want to have a different namespace for the
# version of buf installed from source
# CACHE_GOBIN takes precedence over CACHE_BIN in PATH
CACHE_GOBIN := $(CACHE)/gobin
# CACHE_GOCACHE is where the build cache is stored.
CACHE_GOCACHE := $(CACHE)/gocache

ifeq ($(UNAME_OS),Darwin)
OPEN_CMD := open
endif
ifeq ($(UNAME_OS),Linux)
OPEN_CMD := xdg-open
endif

ifeq ($(UNAME_OS),Darwin)
SED_I := sed -i ''
endif
ifeq ($(UNAME_OS),Linux)
SED_I := sed -i
endif

# Runtime ALL

# All variables exported here must also be added to env.sh
# via the direnv target below
export GO111MODULE := on
ifdef GOPRIVATE
export GOPRIVATE := $(GOPRIVATE),$(GO_MODULE)
else
export GOPRIVATE := $(GO_MODULE)
endif
export GOPATH := $(abspath $(CACHE_GO))
export GOBIN := $(abspath $(CACHE_GOBIN))
export GOCACHE := $(abspath $(CACHE_GOCACHE))
export GOMODCACHE := $(GOPATH)/pkg/mod

ifdef EXTRAPATH
EXTRAPATH := $(GOBIN):$(abspath $(CACHE_BIN)):$(EXTRAPATH)
else
EXTRAPATH := $(GOBIN):$(abspath $(CACHE_BIN))
endif
export PATH := $(EXTRAPATH):$(PATH)
export DOCKER_BUILDKIT := 1

print-%:
	@echo $($*)

.PHONY: envbackup
envbackup:
	rm -rf "$(ENV_BACKUP_DIR)"
	mkdir -p "$(dir $(ENV_BACKUP_DIR))"
	cp -R "$(ENV_DIR)" "$(ENV_BACKUP_DIR)"

.PHONY: envrestore
envrestore:
	@ if [ ! -d "$(ENV_BACKUP_DIR)" ]; then echo "no backup stored in $(ENV_BACKUP_DIR)"; exit 1; fi
	rm -rf "$(ENV_DIR)"
	cp -R "$(ENV_BACKUP_DIR)" "$(ENV_DIR)"

# All variables set in env.sh by this target need to also be exported
# above in the Runtime ALL section.
.PHONY: direnv
direnv:
	@mkdir -p $(CACHE_ENV)
	@rm -f $(CACHE_ENV)/env.sh
	@echo 'export CACHE="$(abspath $(CACHE))"' >> $(CACHE_ENV)/env.sh
	@echo 'export GO111MODULE="$(GO111MODULE)"' >> $(CACHE_ENV)/env.sh
	@echo 'export GOPRIVATE="$(GOPRIVATE)"' >> $(CACHE_ENV)/env.sh
	@echo 'export GOPATH="$(GOPATH)"' >> $(CACHE_ENV)/env.sh
	@echo 'export GOBIN="$(GOBIN)"' >> $(CACHE_ENV)/env.sh
	@echo 'export GOCACHE="$(GOCACHE)"' >> $(CACHE_ENV)/env.sh
	@echo 'export GOMODCACHE="$(GOPATH)/pkg/mod"' >> $(CACHE_ENV)/env.sh
	@echo 'export PATH="$(EXTRAPATH):$${PATH}"' >> $(CACHE_ENV)/env.sh
	@echo 'export DOCKER_BUILDKIT=1' >> $(CACHE_ENV)/env.sh
	@echo '[ -f "$(abspath $(ENV_SH))" ] && . "$(abspath $(ENV_SH))"' >> $(CACHE_ENV)/env.sh
	@echo $(CACHE_ENV)/env.sh

.PHONY: clean
clean:
	git clean -xdf -e /$(ENV_DIR)/

.PHONY: cleancache
cleancache:
	rm -rf $(CACHE_BASE)

.PHONY: nuke
nuke: clean
	sudo rm -rf $(CACHE_GO)/pkg/mod
	rm -rf $(CACHE_BASE)

.PHONY: dockerdeps
dockerdeps::

.PHONY: deps
deps:: dockerdeps

.PHONY: preinstallgenerate
preinstallgenerate::

.PHONY: pregenerate
pregenerate::

.PHONY: prepostgenerate
prepostgenerate::

.PHONY: postprepostgenerate
postprepostgenerate::

.PHONY: postgenerate
postgenerate::

.PHONY: licensegenerate
licensegenerate::

.PHONY: generate
generate:
	@$(MAKE) preinstallgenerate
	@$(MAKE) pregenerate
	@$(MAKE) prepostgenerate
	@$(MAKE) postprepostgenerate
	@$(MAKE) postgenerate
	@$(MAKE) licensegenerate

.PHONY: checknodiffgenerated
checknodiffgenerated:
	@ if [ -d .git ]; then \
			$(MAKE) __checknodiffgeneratedinternal; \
		else \
			echo "skipping make checknodiffgenerated due to no .git repository" >&2; \
		fi

.PHONY: preupgrade
preupgrade::

.PHONY: postupgrade
postupgrade::

.PHONY: upgrade
upgrade:
	@$(MAKE) preupgrade
	@$(MAKE) generate
	@$(MAKE) postupgrade

.PHONY: upgradenopost
upgradenopost:
	@$(MAKE) preupgrade
	@$(MAKE) generate

.PHONY: copyfrommakego
copyfrommakego:
	@rm -rf $(TMP)/makego
	@mkdir -p $(TMP)
	git clone $(MAKEGO_REMOTE) $(TMP)/makego
	rm -rf $(MAKEGO)
	cp -R $(TMP)/makego/make/go $(MAKEGO)
	@rm -rf $(TMP)/makego

.PHONY: copytomakego
copytomakego:
	@rm -rf $(TMP)/makego
	@mkdir -p $(TMP)
	git clone $(MAKEGO_REMOTE) $(TMP)/makego
	rm -rf $(TMP)/makego/make/go
	cp -R $(MAKEGO) $(TMP)/makego/make/go
	bash $(MAKEGO)/scripts/pushall.bash $(TMP)/makego

.PHONY: initmakego
initmakego::

.PHONY: updategitignores
updategitignores:
	@rm -f .gitignore
	@echo '# Autogenerated by makego. DO NOT EDIT.' > .gitignore
	@$(foreach file_ignore,$(sort $(FILE_IGNORES)),echo /$(file_ignore) >> .gitignore || exit 1; )
	@$(foreach git_file_ignore,$(sort $(GIT_FILE_IGNORES)),echo $(git_file_ignore) >> .gitignore || exit 1; )

pregenerate:: updategitignores

.PHONY: __checknodiffgeneratedinternal
__checknodiffgeneratedinternal:
	bash $(MAKEGO)/scripts/checknodiffgenerated.bash $(MAKE) generate
