# go makefile

program := $(notdir $(shell pwd))
version := $(shell cat VERSION)

include make/common.mk

default: build

build: $(binary)

$(binary): fmt
	fix go build . ./...
	go build

fmt: go.sum
	fix go fmt . ./...

go.mod:
	go mod init

go.sum: go.mod
	go mod tidy

install: build
	go install

test: fmt
	go test -v -failfast . ./...

debug: fmt
	go test -v -failfast -count=1 -run $(test) . ./...

release:
	$(gitclean)
	@$(if $(update),gh release delete -y v$(version),)
	gh release create v$(version) --notes "v$(version)"
	$(MAKE) build

dist: dist/$(release_binary)

dist/$(release_binary): $(binary)
	mkdir -p dist
	scp $< $(dist_target)/$(release_binary)
	scp $< $(dist_target)/$(dist_binary)
	cp $< $@

release-upload: dist
	cd dist; gh release upload $(latest_release) $(release_binary) $(CLOBBER)

update-modules:
	@echo checking dependencies for updated versions 
	@$(foreach module,$(rstms_modules),go get $(module)@$(call latest_module_release,$(module)) ;)
	curl -Lso .proxy https://raw.githubusercontent.com/rstms/go-common/master/proxy_common_go
	@$(foreach s,$(common_go),sed <.proxy >$(s) 's/^package cmd/package $(lastword $(subst /, ,$(dir $(s))))/' ;)
	-rm .proxy

clean:
	rm -f $(binary) *.core 
	go clean
	rm -rf dist && mkdir dist

sterile: clean
	go clean -i || true
	go clean
	go clean -cache
	go clean -modcache
	rm -f go.mod go.sum

show-vars:
	@$(foreach var,$(all_variables),echo $(var)=$($(var));)
