regenerate:
	find ./test -type f -name *.proto -exec \
		protoc \
			--proto_path=..:. \
			-I=${GOPATH}/src/ \
			--go-sqlmap_out=Msqlgen/sqlgen.proto=github.com/roderm/protoc-gen-go-sqlmap/sqlgen:. \
		{} \;

install: dependencies proto
	buf generate
	go install