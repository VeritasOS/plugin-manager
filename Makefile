# Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.DEFAULT_GOAL := help
.PHONY: help
help:	## Display this help message.
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

TOP=$(CURDIR)
include $(TOP)/Makefile.conf

# Go build related variables
GOSRC=$(TOP)

# Set GOBIN to where binaries get picked up while creating RPM/ISO.
GOBIN?=$(TOP)/bin
GOCOVERDIR=$(GOSRC)/cover
GOTOOLSBIN=$(TOP)/tools/go/
PROTOBUF_PATH=$(TOP)/tools/protobuf/

.SILENT:

.PHONY: all
all: clean analyze build test

.PHONY: clean
clean: 	## Clean Plugin Manager go build & test artifacts
	@echo "Cleaning Plugin Manager Go binaries...";
	export GOBIN=$(GOBIN); \
	cd $(GOSRC); \
	go clean -i -mod=vendor ./...;
	@echo "Cleaning Go test artifacts... ";
	-@rm $(GOSRC)/{,.}*{dot,html,log,svg,xml};
	-@rm $(GOSRC)/cmd/pm/{,.}*{dot,log,svg};
	-@rm -rf $(GOSRC)/{,cmd/pm/}plugins/
	-@rm -rf $(GOCOVERDIR);

.PHONY: build
build: compile-proto ## Build source code
	# Since go build determines and build only updated sources, no need to run clean all go binaries
	@echo "Building Plugin Manager Go binaries...";
	export GOBIN=$(GOBIN); \
	cd $(GOSRC); \
	go install -ldflags "-X main.buildDate=`date -u +%Y%m%d.%H%M%S`" -mod=vendor -v ./...; \
	ret=$$?; \
	if [ $${ret} -ne 0 ]; then \
		@echo "Failed to build Plugin Manager Go binaries."; \
		exit 1; \
	fi


.PHONY: analyze
analyze: gofmt golint govet go-race  ## Analyze source code for different errors through gofmt, golint, govet, go-race

.PHONY: golint
golint:	## Run golint
	@echo Checking Plugin Manager Go code for lint errors...
	$(GOTOOLSBIN)/golint -set_exit_status `cd $(GOSRC); go list -mod=vendor  -f '{{.Dir}}' ./...`

.PHONY: gofmt
gofmt:	## Run gofmt
	@echo Checking Go code for format errors...
	fmterrs=`gofmt -l . | grep -v vendor/ 2>&1`; \
	if [ "$$fmterrs" ]; then \
		echo "gofmt must be run on the following files:"; \
		echo "$$fmterrs"; \
		exit 1; \
	fi

.PHONY: govet
govet:	## Run go vet
	@echo Vetting Plugin Manager Go code for errors...
	cd $(GOSRC); \
	go vet -mod=vendor -all ./...

.PHONY: test
test:  	## Run all tests
	echo "Running Plugin Manager Go Unit Tests...";
	mkdir -p $(GOCOVERDIR);
	export INTEG_TEST_BIN=$(GOSRC); \
	export PM_CONF_FILE=$(GOSRC)/sample/pm.config.yaml; \
	export INTEGRATION_TEST=START; \
	export GOCOVERDIR=$(GOCOVERDIR); \
	cd $(GOSRC); \
	test_failed=0; \
	d=pm; \
	go test -mod=vendor -v --cover -covermode=count -coverprofile=$(GOCOVERDIR)/$${d}.out ./... | \
		$(GOTOOLSBIN)/go-junit-report > TEST-$${d}.xml; \
	ret=$${PIPESTATUS[0]}; \
	if [ $${ret} -ne 0 ]; then \
		echo "Go unit test failed for $${d}."; \
		test_failed=1; \
	fi ; \
	awk -f $(TOP)/tools/gocoverage-collate.awk $(GOCOVERDIR)/* > $(GOCOVERDIR)/cover.out; \
	go tool cover -html=$(GOCOVERDIR)/cover.out -o go-coverage-$${d}.html; \
	$(GOTOOLSBIN)/gocov convert $(GOCOVERDIR)/cover.out | $(GOTOOLSBIN)/gocov-xml > go-coverage-$${d}.xml; \
	rm -rf $(GOCOVERDIR)/*; \
	export INTEGRATION_TEST=DONE; \
	if [ $${test_failed} -ne 0 ]; then \
		echo "Go unit tests failed."; \
		exit 1; \
	fi

.PHONY: go-race
go-race: 	## Run Go tests with race detector enabled
	echo "Checking Go code for race conditions...";
	# NOTE: COVER directory should be present, along with INTEGRATION_TEST
	# 	value being set to "START" for integ_test.go to succeed.
	mkdir -p $(GOCOVERDIR);
	export INTEGRATION_TEST=START; \
	export INTEG_TEST_BIN=$(GOSRC); \
	export GOCOVERDIR=$(GOCOVERDIR); \
	cd $(GOSRC); \
	export PM_CONF_FILE=$(GOSRC)/sample/pm.config.yaml; \
	go test -mod=vendor -v -race ./...;

.NOTPARALLEL:

.PHONY: update-go-tools
update-go-tools:
	export GOBIN=$(GOTOOLSBIN); \
	go install github.com/axw/gocov/gocov@latest; \
	go install github.com/AlekSi/gocov-xml@latest; \
	go install github.com/matm/gocov-html/cmd/gocov-html@latest; \
	go install github.com/jstemmer/go-junit-report/v2@latest; \
	go get -u golang.org/x/lint/golint;

.PHONY: install-go
install-go:
	wget -c https://go.dev/dl/go1.23.1.linux-amd64.tar.gz -P /tmp
	ret=$$?; \
	if [ $${ret} -ne 0 ]; then \
		echo "Failed to download go install files. Return: $${d}."; \
		exit 1; \
	fi ;
	rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go*.linux-amd64.tar.gz
	export PATH=/usr/local/go/bin:$PATH

.PHONY: install-proto-deps
install-proto-deps:
	wget -c https://github.com/protocolbuffers/protobuf/releases/download/v28.0/protoc-28.0-linux-x86_64.zip -P $(PROTOBUF_PATH)
	ret=$$?; \
	if [ $${ret} -ne 0 ]; then \
		echo "Failed to download protobuf protoc. Return: $${d}."; \
		exit 1; \
	fi ;
	cd $(PROTOBUF_PATH); unzip protoc-*-linux-x86_64.zip;
	ret=$$?; \
	if [ $${ret} -ne 0 ]; then \
		echo "Failed to unzip protoc*.zip. Return: $${d}."; \
		exit 1; \
	fi ;
	export GOBIN=$(PROTOBUF_PATH)/bin; \
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	ret=$$?; \
	if [ $${ret} -ne 0 ]; then \
		echo "Failed to install protoc-gen-go@v1.28. Return: $${d}."; \
		exit 1; \
	fi ; \
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	ret=$$?; \
	if [ $${ret} -ne 0 ]; then \
		echo "Failed to install protoc-gen-go-grpc@v1.3. Return: $${d}."; \
		exit 1; \
	fi ;

.PHONY: compile-proto
compile-proto:
	export PATH=$(PROTOBUF_PATH)/bin/:$(PATH); \
	protoc -I ./proto -I $(PROTOBUF_PATH)/include/google/protobuf/ --go_out=. --go_opt=module=github.com/VeritasOS/plugin-manager --go-grpc_out=./pluginmanager --go-grpc_opt=paths=source_relative proto/*.proto
	ret=$$?; \
	if [ $${ret} -ne 0 ]; then \
		echo "Failed to compile proto files. Return: $${d}."; \
		exit 1; \
	fi ; \

.PHONY: clean-proto
clean-proto:
	-@rm -rf pluginmanager/*.pb.go;
