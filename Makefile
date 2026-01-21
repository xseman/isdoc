OUTPUT_DIR ?= bin
LDFLAGS := -s -w

.PHONY: build
build:
	@mkdir -p $(OUTPUT_DIR)
	@CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/isdoc ./cmd/isdoc

.PHONY: build-ffi
build-ffi:
	@mkdir -p $(OUTPUT_DIR)
	@CGO_ENABLED=1 go build -buildmode=c-shared -o $(OUTPUT_DIR)/libisdoc ./cmd/libisdoc

.PHONY: test
test:
	@go test -v -race ./...

.PHONY: cover
cover:
	@go test -race -coverprofile=coverage.out -covermode=atomic ./...

.PHONY: clean
clean:
	@rm -rf $(OUTPUT_DIR)
	@rm -f coverage.out
