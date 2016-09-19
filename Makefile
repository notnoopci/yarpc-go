# Paths besides auto-detected generated files that should be excluded from
# lint results.
LINT_EXCLUDES_EXTRAS =

##############################################################################
export GO15VENDOREXPERIMENT=1

PACKAGES := $(shell glide novendor)

GO_FILES := $(shell \
	find . '(' -path '*/.*' -o -path './vendor' ')' -prune \
	-o -name '*.go' -print | cut -b3-)

# Files whose first line contains "Code generated by" are generated.
GENERATED_GO_FILES := $(shell \
	find $(GO_FILES) \
	-exec sh -c 'head -n30 {} | grep "Code generated by\|\(Autogenerated\|Automatically generated\) by\|@generated" >/dev/null' \; \
	-print)

LINT_EXCLUDES := $(GENERATED_GO_FILES) $(LINT_EXCLUDES_EXTRAS)

# Pipe lint output into this to filter out ignored files.
FILTER_LINT := grep -v $(patsubst %,-e %, $(LINT_EXCLUDES))

CHANGELOG_VERSION = $(shell grep '^v[0-9]' CHANGELOG.md | head -n1 | cut -d' ' -f1)
INTHECODE_VERSION = $(shell perl -ne '/^const Version.*"([^"]+)".*$$/ && print "v$$1\n"' version.go)
##############################################################################

.PHONY: build
build:
	go build $(PACKAGES)

.PHONY: lint
lint:
	$(eval FMT_LOG := $(shell mktemp -t gofmt.XXXXX))
	@gofmt -e -s -l $(GO_FILES) | $(FILTER_LINT) > $(FMT_LOG) || true
	@[ ! -s "$(FMT_LOG)" ] || (echo "gofmt failed:" | cat - $(FMT_LOG) && false)

	$(eval VET_LOG := $(shell mktemp -t govet.XXXXX))
	@go vet $(PACKAGES) 2>&1 | grep -v '^exit status' | $(FILTER_LINT) > $(VET_LOG) || true
	@[ ! -s "$(VET_LOG)" ] || (echo "govet failed:" | cat - $(VET_LOG) && false)

	$(eval LINT_LOG := $(shell mktemp -t golint.XXXXX))
	@cat /dev/null > $(LINT_LOG)
	@$(foreach pkg, $(PACKAGES), golint $(pkg) | $(FILTER_LINT) >> $(LINT_LOG) || true;)
	@[ ! -s "$(LINT_LOG)" ] || (echo "golint failed:" | cat - $(LINT_LOG) && false)

.PHONY: install
install:
	glide --version || go get github.com/Masterminds/glide
	glide install


.PHONY: test
test: verify_version
	go test $(PACKAGES)


.PHONY: cover
cover:
	./scripts/cover.sh $(shell go list $(PACKAGES))
	go tool cover -html=cover.out -o cover.html


# This is not part of the regular test target because we don't want to slow it
# down.
.PHONY: test-examples
test-examples:
	make -C examples


.PHONY: crossdock
crossdock:
	docker-compose kill go
	docker-compose rm -f go
	docker-compose build go
	docker-compose run crossdock


.PHONY: crossdock-fresh
crossdock-fresh: install
	docker-compose kill
	docker-compose rm --force
	docker-compose pull
	docker-compose build
	docker-compose run crossdock

.PHONY: test_ci
test_ci: verify_version
	./scripts/cover.sh $(shell go list $(PACKAGES))

.PHONY: verify_version
verify_version:
	@if [ "$(INTHECODE_VERSION)" = "$(CHANGELOG_VERSION)" ]; then \
		echo "yarpc-go: $(CHANGELOG_VERSION)"; \
	elif [ "$(INTHECODE_VERSION)" = "$(CHANGELOG_VERSION)-dev" ]; then \
		echo "yarpc-go (development): $(INTHECODE_VERSION)"; \
	else \
		echo "Version number in version.go does not match CHANGELOG.md"; \
		echo "version.go: $(INTHECODE_VERSION)"; \
		echo "CHANGELOG : $(CHANGELOG_VERSION)"; \
		exit 1; \
	fi
