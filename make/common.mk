# go makefile common include


#
# detect os
#
os := $(if $(SYSTEMROOT),windows,$(shell uname | tr A-Z a-z))

#
# os specific
#

ifeq ($(os),windows)
  hostname := $(shell hostname)
  fqdn := "$(hostname).$(shell powershell -Command '(Get-WmiObject Win32_NetworkAdapterConfiguration | Where-Object { $$_.IPEnabled -eq $$true }).DNSDomain')"
  os_version := $(lastword $(subst ],,$(shell cmd /c ver)))
  arch := $(shell echo $$PROCESSOR_ARCHITECTURE | tr A-Z a-z)
  windows := 1	
  USER := $(USERNAME)
  HOME := /c/Users/$(USER)
endif

ifeq ($(os),linux)
  hostname := $(shell hostname -s)
  fqdn := $(shell hostname --fqdn)
  os_version := $(firstword $(subst +, ,$(subst -, ,$(shell uname -r))))
  arch := $(shell uname -m)
  linux := 1
endif

ifeq ($(os),openbsd)
  hostname := $(shell hostname -s)
  fqdn := $(shell hostname)
  os_version := $(shell uname -r)
  arch := $(shell uname -m)
  openbsd := 1
endif

binary_extension := $(if $(windows),.exe,)
binary := $(program)$(binary_extension)

#
# module versions
#
rstms_modules = $(shell awk <go.mod '/^module/{next} /rstms/{print $$1}')
common_go = $(wildcard */common.go) $(wildcard */*/common.go)
latest_module_release = $(shell gh --repo $(1) release list --json tagName --jq '.[0].tagName')

#
# release
#
org := rstms
latest_release = $(call latest_module_release,$(org)/$(program))
gitclean = $(if $(shell git status --porcelain),$(error git status is dirty),$(info git status is clean))
release_binary = $(program)-v$(version)-$(os)-$(os_version)-$(arch)$(binary_extension)
dist_target := vega.rstms.net:rstms/dist
dist_binary := $(program)-latest-$(os)-$(os_version)-$(arch)$(binary_extension)

#
# configuration
#
config_dir := $(if $(windows),$(HOME)/AppData/Roaming/$(program),$(HOME)/.config/$(program))
cache_dir := $(if $(windows),$(HOME)/AppData/Local/$(program),$(HOME)/.cache/$(program))

#
# diagnostics
#
all_variables = \
    HOME USER program version org \
    os arch os_version hostname fqdn windows openbsd linux \
    binary_extension binary \
    rstms_modules common_go \
    latest_release release_binary dist_target dist_binary \
    config_dir cache_dir
