package plain

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func (b *SqlGenerator) GenerateOrm() (*pluginpb.CodeGeneratorResponse, error) {
	genFileMap := make(map[string]*protogen.GeneratedFile)

	for _, protoFile := range b.plugin.Files {
		fileName := protoFile.GeneratedFilenamePrefix + ".pb.gorm.go"
		g := b.plugin.NewGeneratedFile(fileName, ".")
		genFileMap[fileName] = g

		b.currentPackage = protoFile.GoImportPath.String()

		// first traverse: preload the messages
		for _, message := range protoFile.Messages {
			if message.Desc.IsMapEntry() {
				continue
			}

			typeName := string(message.Desc.Name())
			b.messages[typeName] = struct{}{}
		}

	}

	for _, protoFile := range b.plugin.Files {
		// generate actual code
		fileName := protoFile.GeneratedFilenamePrefix + ".pb.gorm.go"
		g, ok := genFileMap[fileName]
		if !ok {
			panic("generated file should be present")
		}

		if !protoFile.Generate {
			g.Skip()
			continue
		}

		g.P("package ", protoFile.GoPackageName)

		// for _, message := range protoFile.Messages {
		// 	if isOrmable(message) {
		// 		b.generateOrmable(g, message)
		// 		b.generateTableNameFunctions(g, message)
		// 		b.generateConvertFunctions(g, message)
		// 		b.generateHookInterfaces(g, message)
		// 	}
		// }

		// b.generateDefaultHandlers(protoFile, g)
		// b.generateDefaultServer(protoFile, g)
	}

	return b.plugin.Response(), nil
}
