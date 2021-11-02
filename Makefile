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
	go mod vendor
	git clone https://github.com/protocolbuffers/protobuf.git vendor/github.com/protocolbuffers/protobuf
	go test -run=.

install:
	buf generate
	go install
