
GOGO_PATH=${GOPATH}/src/github.com/gogo/protobuf
.PHONY: proto
proto: 
	# check how to do it with GO111MODULE=on
	# IMPORT_PATH=
	# $(eval IMPORT_PATH=${IMPORT_PATH}:${GOPATH}/src/)
	# $(eval IMPORT_PATH=${IMPORT_PATH}:${GOGO_PATH}/)
	# $(eval IMPORT_PATH=${IMPORT_PATH}:${GOGO_PATH}/protobuf/)
	# GO111MODULE=off go get google.golang.org/grpc/cmd/protoc-gen-go-grpc;
	# failes cause there aren't any go files, so we pass true
	# go install github.com/gogo/protobuf/protoc-gen-gogo | true;
	# GO111MODULE=off go get github.com/gogo/protobuf | true; 
	protoc \
		--proto_path=. \
		--gogo_out=Mgogo/protobuf/google/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. \
		sqlgen/sqlgen.proto

regenerate:
	find ./test -type f -name *.proto -exec \
		protoc --go-sqlmap_out=Msqlgen/sqlgen.proto=github.com/roderm/protoc-gen-go-sqlmap/sqlgen:. \
		--proto_path=..:. \
		{} \;

install: proto
	go install
