## A Pröblem:
* Proto Import: `github.com/gogo/protobuf/protobuf/google/protobuf/descriptor.proto`
* actual go-package: `github.com/gogo/protobuf/protoc-gen-gogo/descriptor`
=> fixed with `gofast_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:.`