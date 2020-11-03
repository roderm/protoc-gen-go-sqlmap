
GOGO_PATH=${GOPATH}/src/github.com/gogo/protobuf
.PHONY: proto
proto: 
	IMPORT_PATH=
	$(eval IMPORT_PATH=${IMPORT_PATH}:${GOPATH}/src/)
	$(eval IMPORT_PATH=${IMPORT_PATH}:${GOGO_PATH}/)
	$(eval IMPORT_PATH=${IMPORT_PATH}:${GOGO_PATH}/protobuf/)
	# $(eval IMPORT_PATH=${IMPORT_PATH}:${GOPATH}/src/github.com/roderm/protoc-gen-go-sqlmap/sqlgen/)
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc;
	go install github.com/gogo/protobuf/protoc-gen-gogo;
	find ./ -type f -name *.proto -exec \
		protoc \
			--proto_path=${IMPORT_PATH}:. \
			--gogo_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. \
		{} \;

install: proto
	go install

.PHONY: test
test:
	protoc \
		--proto_path=${GOPATH}/src/:${GOPATH}/src/github.com/gogo/protobuf/:${GOPATH}/src/github.com/gogo/protobuf/protobuf:. \
		--go-sqlmap_out=. \
		test/test1.proto