
GOGO_PATH=${GOPATH}/src/github.com/gogo/protobuf
dependencies:
	GO111MODULE=off go get github.com/gogo/protobuf | true;
	go install github.com/gogo/protobuf/protoc-gen-gogo | true;

.PHONY: proto 
proto: 
	protoc \
		-I=${GOPATH}/src \
		-I=. \
		--gogo_out=Mgithub.com/gogo/protobuf/protobuf/google/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:${GOPATH}/src/ \
		sqlgen/sqlgen.proto
	protoc -I=. --gogo_out=${GOPATH}/src/ lib/proto/timestamptz/timestamptz.proto

regenerate:
	find ./test -type f -name *.proto -exec \
		protoc \
			--proto_path=..:. \
			-I=${GOPATH}/src/ \
			--go-sqlmap_out=Msqlgen/sqlgen.proto=github.com/roderm/protoc-gen-go-sqlmap/sqlgen:. \
		{} \;

install: dependencies proto
	go install
