
GOGO_PATH=${GOPATH}/src/github.com/gogo/protobuf
.PHONY: proto
proto: 
	IMPORT_PATH=
	$(eval IMPORT_PATH=${IMPORT_PATH}:${GOPATH}/src/)
	$(eval IMPORT_PATH=${IMPORT_PATH}:${GOGO_PATH}/)
	$(eval IMPORT_PATH=${IMPORT_PATH}:${GOGO_PATH}/protobuf/)
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc;
	go install github.com/gogo/protobuf/protoc-gen-gogo;
	find ./ -type f -name *.proto -exec \
		protoc \
			--proto_path=${IMPORT_PATH}:. \
			--gofast_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. \
		{} \;

install:
	go get github.com/gogo/protobuf
	cd ${GOGO_PATH}/protobuf