
GOGO_PATH=${GOPATH}/src/github.com/gogo/protobuf
.PHONY: proto
proto: 
	# check how to do it with GO111MODULE=on
	IMPORT_PATH=
	$(eval IMPORT_PATH=${IMPORT_PATH}:${GOPATH}/src/)
	$(eval IMPORT_PATH=${IMPORT_PATH}:${GOGO_PATH}/)
	$(eval IMPORT_PATH=${IMPORT_PATH}:${GOGO_PATH}/protobuf/)
	GO111MODULE=off go get google.golang.org/grpc/cmd/protoc-gen-go-grpc;
	# failes cause there aren't any go files, so we pass true
	GO111MODULE=off go get github.com/gogo/protobuf | true; 
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