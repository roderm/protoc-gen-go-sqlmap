regenerate:
	find ./test -type f -name *.proto -exec \
		protoc \
			--proto_path=. \
			-I=${GOPATH}/src/ \
			--go-sqlmap_out=Mproto/sqlgen/v1/sqlgen.proto=github.com/roderm/protoc-gen-go-sqlmap/lib/go/proto/sqlgen,paths=source_relative:. \
		{} \;

install:
	buf generate
	go install
