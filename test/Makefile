test:
	go test
	make -C testdata test

regenerate:
	find ./ -type f -name *.proto -exec \
		protoc --go-sqlmap_out=Msqlgen/sqlgen.proto=github.com/roderm/protoc-gen-go-sqlmap/sqlgen:golden/ \
		--proto_path=${GOPATH}/src/github.com/gogo/protobuf/protobuf:..:. \
	{} \;