.PHONY: regenerate
regenerate:
	go mod vendor;
	git clone https://github.com/protocolbuffers/protobuf.git vendor/github.com/protocolbuffers/protobuf;
	find ./test -type f -name *.proto -exec \
		protoc \
			--proto_path=. \
			-I./vendor/github.com/protocolbuffers/protobuf/src \
			--go-sqlmap_out=paths=source_relative:. \
		{} \;

.PHONY: test
test:
	go build ./...
	go vet ./...
	go test ./...

# Opt-in: starts a throwaway PostgreSQL container, so it needs docker as well
# as protoc and protoc-gen-go.
.PHONY: test-e2e
test-e2e:
	SQLMAP_E2E=1 go test ./pkg/generator/sqlmap/ -run TestE2E -v

# Rewrites testdata/*.sqlmap.go.golden from the current generator output.
.PHONY: golden
golden:
	go test ./pkg/generator/sqlmap/ -update

.PHONY: install
install:
	buf generate
	go install ./cmd/protoc-gen-go-sqlmap/
