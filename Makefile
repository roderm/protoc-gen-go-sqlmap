regenerate:
	find ./test -type f -name *.proto -exec \
		protoc \
			--proto_path=..:. \
			-I=${GOPATH}/src/ \
			-I=${GOPATH}/src/github.com/protocolbuffers/protobuf/src/ \
			--go-sqlmap_out=Msqlgen/sqlgen.proto=github.com/roderm/protoc-gen-go-sqlmap/sqlgen:. \
		{} \;

install:
	buf generate
	go install