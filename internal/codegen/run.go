package codegen

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/LLKennedy/protoc-gen-tsjson/internal/version"
	"github.com/LLKennedy/protoc-gen-tsjson/tsjsonpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var protocVersion = "unknown"

var packageReplacement = regexp.MustCompile(`\.([a-zA-Z0-9_]+)\.(.*)`)

// At time of writing, the only feature that can be marked as supported is restoring the "optional" keyword to proto3, still an experimental feature that is out of scope for this project.
var support uint64 = uint64(pluginpb.CodeGeneratorResponse_FEATURE_NONE)

// Run performs code generation on the input data
func Run(request *pluginpb.CodeGeneratorRequest) (response *pluginpb.CodeGeneratorResponse) {
	defer func() {
		if r := recover(); r != nil {
			response = &pluginpb.CodeGeneratorResponse{
				SupportedFeatures: &support,
				Error:             proto.String(fmt.Sprintf("caught panic in protoc-gen-tsjson: %v", r)),
			}
		}
	}()
	// Set runtime version of protoc
	protocVersion = version.FormatProtocVersion(request.GetCompilerVersion())
	// Create a basic response with our feature support (none, see above)
	response = &pluginpb.CodeGeneratorResponse{
		SupportedFeatures: &support,
	}
	// Make sure the request actually exists as a safeguard
	if request == nil {
		response.Error = proto.String("cannot generate from nil input")
		return
	}
	// Generate the files (do the thing)
	generatedFiles, err := generateAllFiles(request)
	if err != nil {
		// It didn't work, ignore any data we generated and only return the error
		response.Error = proto.String(fmt.Sprintf("failed to generate files: %v", err))
		return
	}
	// It worked, set the response data
	response.File = generatedFiles
	return
}

const googlePrefix = "google."

type exportDetails struct {
	npmPackage   string
	importPath   string
	protoPackage string
}

// Naive approach to codegen, creates output files for every message/service in every linked file, not just the parts depended on by the "to generate" files
func generateAllFiles(request *pluginpb.CodeGeneratorRequest) (outfiles []*pluginpb.CodeGeneratorResponse_File, err error) {
	var out *pluginpb.CodeGeneratorResponse_File
	// Map of file names to input paths
	var exportMap map[string]exportDetails
	// Map of package names to type names to import details
	var typeMap map[string]map[string]exportDetails
	exportMap, typeMap, err = buildImportsAndTypes(request.GetProtoFile())
	if err != nil {
		return nil, err
	}
	for _, file := range request.GetProtoFile() {
		for _, toGen := range request.GetFileToGenerate() {
			if file.GetName() == toGen {
				out, err = generateFullFile(file, exportMap, typeMap)
				if err != nil {
					return
				}
				outfiles = append(outfiles, out)
				break
			}
		}
	}
	return
}

func buildImportsAndTypes(files []*descriptorpb.FileDescriptorProto) (exportMap map[string]exportDetails, typeMap map[string]map[string]exportDetails, err error) {
	// Map of file names to input paths
	exportMap = make(map[string]exportDetails, len(files))
	// Map of package names to type names to import details
	typeMap = make(map[string]map[string]exportDetails, len(files)) // Length here is just a starting value, not expected to be accurate
	// Map of
	// Check all files except google ones have both npm_package and import_path options set
	for _, file := range files {
		pkgName := file.GetPackage()
		if len(pkgName) >= len(googlePrefix) && pkgName[:len(googlePrefix)] == googlePrefix {
			// Google files are allowed to not have the options, we handle them differently
			continue
		}
		npmPackage, ok := proto.GetExtension(file.GetOptions(), tsjsonpb.E_NpmPackage).(string)
		if !ok || npmPackage == "" {
			return nil, nil, fmt.Errorf("all imported files must specify the option (tsjson.npm_package), file %s did not", file.GetName())
		}
		importPath, _ := proto.GetExtension(file.GetOptions(), tsjsonpb.E_ImportPath).(string)
		pkg := file.GetPackage()
		details := exportDetails{
			npmPackage:   npmPackage,
			importPath:   importPath,
			protoPackage: pkg,
		}
		exportMap[file.GetName()] = details
		pkgTypes, ok := typeMap[pkg]
		if !ok {
			pkgTypes = make(map[string]exportDetails, len(file.GetEnumType())+len(file.GetMessageType()))
			typeMap[pkg] = pkgTypes
		}
		// Map out type defintions to packages for lookup later
		for _, enum := range file.GetEnumType() {
			parsedName := strings.ReplaceAll(enum.GetName(), ".", "__")
			pkgTypes[parsedName] = details
		}
		for _, msg := range file.GetMessageType() {
			parsedName := strings.ReplaceAll(msg.GetName(), ".", "__")
			pkgTypes[parsedName] = details
			for _, innerMsg := range msg.GetNestedType() {
				innerName := fmt.Sprintf("%s__%s", parsedName, strings.ReplaceAll(innerMsg.GetName(), ".", "__"))
				pkgTypes[innerName] = details
			}
		}
	}
	return exportMap, typeMap, nil
}

func generatePackages(request *pluginpb.CodeGeneratorRequest) (out []*pluginpb.CodeGeneratorResponse_File, pkgMap map[string]string, err error) {
	pkgMap = make(map[string]string)
	packageNames := map[string][]string{}
	for _, file := range request.GetProtoFile() {
		if file.GetSyntax() != "proto3" {
			err = fmt.Errorf("proto3 is the only syntax supported by protoc-gen-tsjson, found %s in %s", file.GetSyntax(), file.GetName())
			return
		}
		pkgName := file.GetPackage()
		if pkgName == "" {
			err = fmt.Errorf("packages are mandatory with protoc-gen-tsjson, %s did not have a package", file.GetName())
			return
		}
		if pkgName == "index" {
			err = fmt.Errorf("for JS/TS language reasons, \"index\" is an invalid package name")
		}
		pkgMap[file.GetName()] = pkgName
		list, _ := packageNames[pkgName]
		list = append(list, file.GetName())
		packageNames[pkgName] = list
	}
	indexFile := &pluginpb.CodeGeneratorResponse_File{
		Name: proto.String("__packages__/index.ts"),
	}
	indexContent := &strings.Builder{}
	for pkgName, importList := range packageNames {
		outFile := &pluginpb.CodeGeneratorResponse_File{
			Name: proto.String(fmt.Sprintf("__packages__/%s.ts", pkgName)),
		}
		content := &strings.Builder{}
		for _, importFile := range importList {
			parsedName := filenameFromProto(importFile)
			content.WriteString(fmt.Sprintf("export * from \"%s\";\n", parsedName.fullWithoutExtension))
		}
		outFile.Content = proto.String(content.String())
		out = append(out, outFile)
		indexContent.WriteString(fmt.Sprintf("export * as %s from \"__packages__/%s\";\n", pkgName, pkgName))
	}
	indexFile.Content = proto.String(indexContent.String())
	out = append(out, indexFile)
	return
}

func generateFullFile(f *descriptorpb.FileDescriptorProto, exportMap map[string]exportDetails, typeMap map[string]map[string]exportDetails) (out *pluginpb.CodeGeneratorResponse_File, err error) {
	if f.GetSyntax() != "proto3" {
		err = fmt.Errorf("proto3 is the only syntax supported by protoc-gen-tsjson, found %s in %s", f.GetSyntax(), f.GetName())
		return
	}
	parsedName := filenameFromProto(f.GetName())
	out = &pluginpb.CodeGeneratorResponse_File{
		Name: proto.String(parsedName.fullWithoutExtension + ".ts"),
	}
	content := &strings.Builder{}
	content.WriteString(getCodeGenmarker(version.GetVersionString(), protocVersion, f.GetName()))
	// Imports
	generateImports(f, content, typeMap)
	// Enums
	generateEnums(f.GetEnumType(), content)
	// Messages
	generateMessages(f.GetMessageType(), content, f.GetPackage())
	// Services
	generateServices(f.GetService(), content)
	// Comments? unclear how to link them back to other elements
	generateComments(f.GetSourceCodeInfo(), content)
	out.Content = proto.String(content.String())
	return
}

func generateImports(f *descriptorpb.FileDescriptorProto, content *strings.Builder, typeMap map[string]map[string]exportDetails) {
	importMap := make(map[string][]string)
	for _, msg := range f.GetMessageType() {
		generateImportsForMessage(f, msg, importMap, content, typeMap)
	}
	for importPath, imports := range importMap {
		fullImportList := &strings.Builder{}
		for i, imp := range imports {
			if i != 0 {
				fullImportList.WriteString(", ")
			}
			fullImportList.WriteString(imp)
		}
		content.WriteString(fmt.Sprintf("import { %s } from \"%s\";\n", fullImportList.String(), importPath))
	}
	content.WriteString("\n")
}

func generateImportsForMessage(f *descriptorpb.FileDescriptorProto, msg *descriptorpb.DescriptorProto, importMap map[string][]string, content *strings.Builder, typeMap map[string]map[string]exportDetails) {
	for _, innerMsg := range msg.GetNestedType() {
		// Recurse
		generateImportsForMessage(f, innerMsg, importMap, content, typeMap)
	}
FIELD_IMPORT_LOOP:
	for _, field := range msg.GetField() {
		typeName := field.GetTypeName()
		if typeName == "" {
			continue
		}
		typeName = strings.TrimLeft(typeName, ".")
		typeNameParts := strings.Split(typeName, ".")
		trueName := typeNameParts[len(typeNameParts)-1]
		pkgName := strings.TrimSuffix(typeName, "."+trueName)
		var importPath string
		ownPkg := f.GetPackage()
		if len(pkgName) >= len(ownPkg) && pkgName[:len(ownPkg)] == ownPkg {
			pkg, ok := typeMap[ownPkg]
			if !ok {
				panic(fmt.Sprintf("failed to find own package %s in imports for file %s", ownPkg, f.GetName()))
			}
			trueName = typeName[len(ownPkg)+1:]
			parsedName := strings.ReplaceAll(trueName, ".", "__")
			// Exclude local messages/enums from import
			for _, msg2 := range f.GetMessageType() {
				if msg2.GetName() == trueName {
					continue FIELD_IMPORT_LOOP
				}
				for _, innerMsg := range msg2.GetNestedType() {
					if msg2.GetName()+"."+innerMsg.GetName() == trueName {
						continue FIELD_IMPORT_LOOP
					}
				}
			}
			details, ok := pkg[parsedName]
			if !ok {
				panic(fmt.Sprintf("failed to find type %s in exports for package %s in file %s", trueName, pkgName, f.GetName()))
			}
			importPath = details.importPath
		} else if pkgName == "google.protobuf" {
			// FIXME: import pre-defined google types from @llkennedy/tsjson here
			continue
		} else {
			log.Println(pkgName + " vs " + ownPkg)
			pkg, ok := typeMap[pkgName]
			if !ok {
				panic(fmt.Sprintf("failed to find package %s in imports for file %s", pkgName, f.GetName()))
			}
			details, ok := pkg[trueName]
			if !ok {
				panic(fmt.Sprintf("failed to find type %s in exports for package %s in file %s", trueName, pkgName, f.GetName()))
			}
			importPath = fmt.Sprintf("%s/%s", details.npmPackage, details.importPath)
		}
		imports, _ := importMap[importPath]
		imports = append(imports, trueName)
		importMap[importPath] = imports
	}
}

func generateEnums(enums []*descriptorpb.EnumDescriptorProto, content *strings.Builder) {
	for _, enum := range enums {
		// TODO: get comment data somehow
		comment := "An enum"
		content.WriteString(fmt.Sprintf("/** %s */\nexport enum %s {\n", comment, enum.GetName()))
		for _, value := range enum.GetValue() {
			// We don't bother stripping the trailing comma on the last enum element because Typescript doesn't care
			// TODO: get comment data somehow
			comment = "An enum value"
			content.WriteString(fmt.Sprintf("	/** %s */\n	%s = %d,\n", comment, value.GetName(), value.GetNumber()))
		}
		content.WriteString("}\n\n")
	}
}

func generateMessages(messages []*descriptorpb.DescriptorProto, content *strings.Builder, pkgName string) {
	for _, message := range messages {
		// TODO: get comment data somehow
		comment := "A message"
		content.WriteString(fmt.Sprintf("/** %s */\nexport class %s extends Object {\n", comment, message.GetName()))
		for _, field := range message.GetField() {
			tsType := getNativeTypeName(field, message, pkgName)
			// FIXME: detect repeated/oneof?
			// TODO: get comment data somehow
			comment = "A field"
			content.WriteString(fmt.Sprintf("	/** %s */\n	public %s?: %s;\n", comment, field.GetJsonName(), tsType))
		}
		content.WriteString("}\n\n")

		for _, nestedType := range message.GetNestedType() {
			if !nestedType.GetOptions().GetMapEntry() {
				// TODO: get comment data somehow
				comment = "A message"
				content.WriteString(fmt.Sprintf("/** %s */\nexport class %s__%s extends Object {\n", comment, message.GetName(), nestedType.GetName()))
				for _, nestedField := range nestedType.GetField() {
					tsType := getNativeTypeName(nestedField, nestedType, pkgName)
					// FIXME: detect repeated/oneof?
					// TODO: get comment data somehow
					comment = "A field"
					content.WriteString(fmt.Sprintf("	/** %s */\n	public %s?: %s;\n", comment, nestedField.GetJsonName(), tsType))
				}
				content.WriteString("}\n\n")
			}
		}
	}
}

func generateServices(services []*descriptorpb.ServiceDescriptorProto, content *strings.Builder) {

}

func generateComments(sourceCodeInfo *descriptorpb.SourceCodeInfo, content *strings.Builder) {

}

func getNativeTypeName(field *descriptorpb.FieldDescriptorProto, message *descriptorpb.DescriptorProto, pkgName string) string {
	switch field.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE,
		descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
		descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		// Javascript only has one number format
		return "number"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "boolean"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "Uint8Array"
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		// TODO: this lookup is not efficient, but it'll do for now. building a map of known types by name as we go would be good
		for _, nestedMessage := range message.GetNestedType() {
			// FIXME: it is possible for this to misfire at least sometimes, though we'll see if it particularly matters
			if nestedMessage.GetOptions().GetMapEntry() && strings.Contains(field.GetTypeName(), nestedMessage.GetName()) {
				keyType := getNativeTypeName(nestedMessage.GetField()[0], nil, pkgName)
				valType := getNativeTypeName(nestedMessage.GetField()[1], nil, pkgName)
				return fmt.Sprintf("Map<%s, %s>", keyType, valType)
			}
		}
		// Not a map
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		typeName := field.GetTypeName()
		matches := packageReplacement.FindStringSubmatch(typeName)
		if len(matches) != 3 {
			panic(fmt.Errorf("type name did not match any valid pattern: %s, found %d instead of 3: %s", typeName, len(matches), matches))
		}
		pkgSection := fmt.Sprintf("packages.%s.", matches[1])
		typeSection := strings.ReplaceAll(matches[2], ".", "__")
		return fmt.Sprintf("%s%s", pkgSection, typeSection)
	default:
		panic(fmt.Errorf("unknown field type: %s", field))
	}
}

func getProtoJSONTypeName(field *descriptorpb.FieldDescriptorProto, nestedTypes []*descriptorpb.DescriptorProto) string {
	panic("not implemented")
}
