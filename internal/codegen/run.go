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
const googleProtobufPrefix = "google.protobuf"

type exportDetails struct {
	npmPackage   string
	importPath   string
	protoPackage string
}

// Naive approach to codegen, creates output files for every message/service in every linked file, not just the parts depended on by the "to generate" files
func generateAllFiles(request *pluginpb.CodeGeneratorRequest) (outfiles []*pluginpb.CodeGeneratorResponse_File, err error) {
	var out *pluginpb.CodeGeneratorResponse_File
	var impexp importsExports
	impexp, err = buildImportsAndTypes(request.GetProtoFile())
	if err != nil {
		return nil, err
	}
	for _, file := range request.GetProtoFile() {
		for _, toGen := range request.GetFileToGenerate() {
			if file.GetName() == toGen {
				out, err = generateFullFile(file, impexp)
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

type importsExports struct {
	exportMap   map[string]exportDetails
	typeMap     map[string]map[string]exportDetails
	fileTypeMap map[string][]string
}

func buildImportsAndTypes(files []*descriptorpb.FileDescriptorProto) (impexp importsExports, err error) {
	// Map of file names to input paths
	impexp.exportMap = make(map[string]exportDetails, len(files))
	// Map of package names to type names to import details
	impexp.typeMap = make(map[string]map[string]exportDetails, len(files)) // Length here is just a starting value, not expected to be accurate
	// Map of file names to type names
	impexp.fileTypeMap = make(map[string][]string, len(files))
	// Map of
	// Check all files except google ones have both npm_package and import_path options set
	for _, file := range files {
		fileName := file.GetName()
		pkgName := file.GetPackage()
		if len(pkgName) >= len(googlePrefix) && pkgName[:len(googlePrefix)] == googlePrefix {
			// Google files are allowed to not have the options, we handle them differently
			continue
		}
		npmPackage, ok := proto.GetExtension(file.GetOptions(), tsjsonpb.E_NpmPackage).(string)
		if !ok || npmPackage == "" {
			err = fmt.Errorf("all imported files must specify the option (tsjson.npm_package), file %s did not", fileName)
		}
		importPath, _ := proto.GetExtension(file.GetOptions(), tsjsonpb.E_ImportPath).(string)
		pkg := file.GetPackage()
		details := exportDetails{
			npmPackage:   npmPackage,
			importPath:   importPath,
			protoPackage: pkg,
		}
		impexp.exportMap[fileName] = details
		pkgTypes, ok := impexp.typeMap[pkg]
		if !ok {
			pkgTypes = make(map[string]exportDetails, len(file.GetEnumType())+len(file.GetMessageType()))
			impexp.typeMap[pkg] = pkgTypes
		}
		// Map out type defintions to packages for lookup later
		for _, enum := range file.GetEnumType() {
			parsedName := strings.ReplaceAll(enum.GetName(), ".", "__")
			pkgTypes[parsedName] = details
			impexp.fileTypeMap[fileName] = append(impexp.fileTypeMap[fileName], parsedName)
		}
		for _, msg := range file.GetMessageType() {
			parsedName := strings.ReplaceAll(msg.GetName(), ".", "__")
			pkgTypes[parsedName] = details
			impexp.fileTypeMap[fileName] = append(impexp.fileTypeMap[fileName], parsedName)
			for _, innerMsg := range msg.GetNestedType() {
				innerName := fmt.Sprintf("%s__%s", parsedName, strings.ReplaceAll(innerMsg.GetName(), ".", "__"))
				pkgTypes[innerName] = details
				impexp.fileTypeMap[fileName] = append(impexp.fileTypeMap[fileName], innerName)
			}
		}
	}
	return impexp, nil
}

func generateFullFile(f *descriptorpb.FileDescriptorProto, impexp importsExports) (out *pluginpb.CodeGeneratorResponse_File, err error) {
	fileName := f.GetName()
	if f.GetSyntax() != "proto3" {
		err = fmt.Errorf("proto3 is the only syntax supported by protoc-gen-tsjson, found %s in %s", f.GetSyntax(), fileName)
		return
	}
	parsedName := filenameFromProto(fileName)
	details, _ := impexp.exportMap[fileName]
	outName := details.importPath
	if outName == "" {
		outName = parsedName.fullWithoutExtension
	}
	out = &pluginpb.CodeGeneratorResponse_File{
		Name: proto.String(outName + ".ts"),
	}
	content := &strings.Builder{}
	content.WriteString(getCodeGenmarker(version.GetVersionString(), protocVersion, fileName))
	// Imports
	generateImports(f, content, impexp)
	// Enums
	generateEnums(f.GetEnumType(), content)
	// Messages
	exports, _ := impexp.fileTypeMap[fileName]
	generateMessages(f.GetMessageType(), content, f.GetPackage(), exports)
	// Services
	generateServices(f.GetService(), content)
	// Comments? unclear how to link them back to other elements
	generateComments(f.GetSourceCodeInfo(), content)
	out.Content = proto.String(content.String())
	return
}

func generateImports(f *descriptorpb.FileDescriptorProto, content *strings.Builder, impexp importsExports) {
	if len(f.GetMessageType()) > 0 {
		// All messages need the common imports
		content.WriteString("import * as tsjson from \"@llkennedy/protoc-gen-tsjson\";\n")
	}
	importMap := make(map[string][]string)
	useGoogle := false
	for _, msg := range f.GetMessageType() {
		useGoogle = useGoogle || generateImportsForMessage(f, msg, importMap, content, impexp)
	}
	if useGoogle {
		content.WriteString("import { google } from \"@llkennedy/protoc-gen-tsjson\";\n")
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

func generateImportsForMessage(f *descriptorpb.FileDescriptorProto, msg *descriptorpb.DescriptorProto, importMap map[string][]string, content *strings.Builder, impexp importsExports) (useGoogle bool) {
	fileName := f.GetName()
	for _, innerMsg := range msg.GetNestedType() {
		// Recurse
		generateImportsForMessage(f, innerMsg, importMap, content, impexp)
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
			pkgName = ownPkg
			pkg, ok := impexp.typeMap[ownPkg]
			if !ok {
				panic(fmt.Sprintf("failed to find own package %s in imports for file %s", ownPkg, fileName))
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
				panic(fmt.Sprintf("failed to find type %s in exports for package %s in file %s", trueName, pkgName, fileName))
			}
			importPath = details.importPath
		} else if pkgName == googleProtobufPrefix {
			useGoogle = true
			continue
		} else {
			pkg, ok := impexp.typeMap[pkgName]
			if !ok {
				panic(fmt.Sprintf("failed to find package %s in imports for file %s", pkgName, fileName))
			}
			details, ok := pkg[trueName]
			if !ok {
				panic(fmt.Sprintf("failed to find type %s in exports for package %s in file %s", trueName, pkgName, fileName))
			}
			importPath = fmt.Sprintf("%s/%s", details.npmPackage, details.importPath)
		}
		imports, _ := importMap[importPath]
		uniqueImports := map[string]struct{}{}
		for _, anImport := range imports {
			uniqueImports[anImport] = struct{}{}
		}
		uniqueImports[fmt.Sprintf("%s as %s__%s", trueName, pkgName, trueName)] = struct{}{}
		imports = []string{}
		for anImport := range uniqueImports {
			imports = append(imports, anImport)
		}
		for _, exp := range impexp.fileTypeMap[fileName] {
			if exp == trueName {
				// This is local, skip
				continue FIELD_IMPORT_LOOP
			}
		}
		importMap[importPath] = imports
	}
	return
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

func generateMessages(messages []*descriptorpb.DescriptorProto, content *strings.Builder, pkgName string, fileExports []string) {
	for _, message := range messages {
		// TODO: get comment data somehow
		comment := "A message"
		generateMessage(message, comment, message.GetName(), pkgName, content, fileExports)
		for _, nestedType := range message.GetNestedType() {
			if !nestedType.GetOptions().GetMapEntry() {
				// TODO: get comment data somehow
				comment = "A message"
				name := fmt.Sprintf("%s__%s", message.GetName(), nestedType.GetName())
				generateMessage(nestedType, comment, name, pkgName, content, fileExports)
			}
		}
	}
}

type mapTypeData struct {
	toProtoJSON string
	parse       string
	keyIsString bool
}

func generateMessage(msg *descriptorpb.DescriptorProto, comment, name, pkgName string, content *strings.Builder, fileExports []string) {
	content.WriteString(fmt.Sprintf("/** %s */\nexport class %s extends Object implements tsjson.ProtoJSONCompatible {\n", comment, name))
	mapTypes := map[string]mapTypeData{}
	for _, nested := range msg.GetNestedType() {
		if nested.GetOptions().GetMapEntry() {
			mapToProtoJSON, mapParse := generateMarshallingStrings(nested.GetField()[1], msg, pkgName, fileExports, mapTypes, "val", `{"value":val}`)
			mapTypes[fmt.Sprintf(".%s.%s.%s", pkgName, msg.GetName(), nested.GetName())] = mapTypeData{
				toProtoJSON: mapToProtoJSON,
				parse:       mapParse,
				keyIsString: nested.GetField()[0].GetType() == descriptorpb.FieldDescriptorProto_TYPE_STRING,
			}
		}
	}
	for _, field := range msg.GetField() {
		if field.GetTypeName() == ".google.protobuf.NullValue" {
			continue
		}
		tsType := getNativeTypeName(field, msg, pkgName, fileExports)
		// FIXME: detect repeated/oneof?
		// TODO: get comment data somehow
		comment = "A field"
		content.WriteString(fmt.Sprintf("	/** %s */\n	public %s?: %s;\n", comment, field.GetJsonName(), tsType))
	}
	protoJSONContent := &strings.Builder{}
	protoJSONContent.WriteString(`		return {
`)
	parseContent := &strings.Builder{}
	parseContent.WriteString(fmt.Sprintf(`		let objData: Object = tsjson.AnyToObject(data);
		let res = new %s();
`, name))
	// Build ToProtoJSON/Parser functions
	for _, field := range msg.GetField() {
		toProtoJSON, parse := generateMarshallingStrings(field, msg, pkgName, fileExports, mapTypes, "this."+field.GetJsonName(), "objData")
		if toProtoJSON == "" && parse == "" {
			// TODO: eventually panic here, this should never happen
			log.Println("skipping field for now: " + field.GetName())
			continue
		}
		protoJSONContent.WriteString(fmt.Sprintf(`			%s: %s,
`, field.GetJsonName(), toProtoJSON))
		parseContent.WriteString(fmt.Sprintf(`		res.%s = await %s;
`, field.GetJsonName(), parse))
	}
	protoJSONContent.WriteString(`		};`)
	parseContent.WriteString(`		return res;`)

	content.WriteString(fmt.Sprintf(`	public ToProtoJSON(): Object {
%s
	}
	public static async Parse(data: any): Promise<%s> {
%s
	}
`, protoJSONContent.String(), name, parseContent.String()))
	content.WriteString("}\n\n")
}

func generateMarshallingStrings(field *descriptorpb.FieldDescriptorProto, msg *descriptorpb.DescriptorProto, pkgName string, fileExports []string, mapTypes map[string]mapTypeData, inputName string, obj string) (toProtoJSON, parse string) {
	label := field.GetLabel()
	switch field.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		switch label {
		case descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL:
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.Bool(%s)`, inputName)
			parse = fmt.Sprintf(`tsjson.Parse.Bool(%s, "%s", "%s")`, obj, field.GetJsonName(), field.GetName())
		case descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.Repeated(tsjson.ToProtoJSON.Bool, %s)`, field.GetJsonName())
			parse = fmt.Sprintf(`tsjson.Parse.Repeated(%s, "%s", "%s", tsjson.PrimitiveParse.Bool())`, obj, field.GetJsonName(), field.GetName())
		}
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		switch label {
		case descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL:
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.Bytes(%s)`, inputName)
			parse = fmt.Sprintf(`tsjson.Parse.Bytes(%s, "%s", "%s")`, obj, field.GetJsonName(), field.GetName())
		case descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.Repeated(tsjson.ToProtoJSON.Bytes, %s)`, inputName)
			parse = fmt.Sprintf(`tsjson.Parse.Repeated(%s, "%s", "%s", tsjson.PrimitiveParse.Bytes())`, obj, field.GetJsonName(), field.GetName())
		}
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, descriptorpb.FieldDescriptorProto_TYPE_FLOAT, descriptorpb.FieldDescriptorProto_TYPE_FIXED32, descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_SFIXED32, descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		switch label {
		case descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL:
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.Number(%s)`, inputName)
			parse = fmt.Sprintf(`tsjson.Parse.Number(%s, "%s", "%s")`, obj, field.GetJsonName(), field.GetName())
		case descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.Repeated(tsjson.ToProtoJSON.Number, %s)`, inputName)
			parse = fmt.Sprintf(`tsjson.Parse.Repeated(%s, "%s", "%s", tsjson.PrimitiveParse.Number())`, obj, field.GetJsonName(), field.GetName())
		}
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64, descriptorpb.FieldDescriptorProto_TYPE_SFIXED64, descriptorpb.FieldDescriptorProto_TYPE_UINT64, descriptorpb.FieldDescriptorProto_TYPE_SINT64, descriptorpb.FieldDescriptorProto_TYPE_INT64:
		switch label {
		case descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL:
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.StringNumber(%s)`, inputName)
			parse = fmt.Sprintf(`tsjson.Parse.Number(%s, "%s", "%s")`, obj, field.GetJsonName(), field.GetName())
		case descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.Repeated(tsjson.ToProtoJSON.StringNumber, %s)`, inputName)
			parse = fmt.Sprintf(`tsjson.Parse.Repeated(%s, "%s", "%s", tsjson.PrimitiveParse.Number())`, obj, field.GetJsonName(), field.GetName())
		}
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		switch label {
		case descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL:
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.String(%s)`, inputName)
			parse = fmt.Sprintf(`tsjson.Parse.String(%s, "%s", "%s")`, obj, field.GetJsonName(), field.GetName())
		case descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.Repeated(tsjson.ToProtoJSON.String, %s)`, inputName)
			parse = fmt.Sprintf(`tsjson.Parse.Repeated(%s, "%s", "%s", tsjson.PrimitiveParse.String())`, obj, field.GetJsonName(), field.GetName())
		}
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		tsType := getNativeTypeName(field, msg, pkgName, fileExports)
		mapStrings, isMap := mapTypes[field.GetTypeName()]
		switch {
		case isMap:
			// TODO: parsers and marshallers for all the different map keys and values, basically nesting this entire switch/case again
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.Map(val => %s, %s)`, mapStrings.toProtoJSON, inputName)
			if mapStrings.keyIsString {
				parse = fmt.Sprintf(`tsjson.Parse.Map(%s, "%s", "%s", %s, %s)`, obj, field.GetJsonName(), field.GetName(), "val => val", mapStrings.parse)
			} else {
				parse = fmt.Sprintf(`tsjson.Parse.Map(%s, "%s", "%s", %s, val => %s)`, obj, field.GetJsonName(), field.GetName(), "JSON.parse", mapStrings.parse)
			}
		case label == descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL:
			toProtoJSON = fmt.Sprintf(`%s?.ToProtoJSON()`, inputName)
			parse = fmt.Sprintf(`tsjson.Parse.Message(%s, "%s", "%s", %s.Parse)`, obj, field.GetJsonName(), field.GetName(), tsType)
		case label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
			toProtoJSON = fmt.Sprintf(`tsjson.ToProtoJSON.Repeated(%s?.ToProtoJSON, %s)`, inputName, inputName)
			parse = fmt.Sprintf(`tsjson.Parse.Repeated(%s, "%s", "%s", %s.Parse)`, obj, field.GetJsonName(), field.GetName(), tsType)
		}
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		// TODO: enums
		switch label {
		case descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL:
		case descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
		}
	}
	return
}

func generateServices(services []*descriptorpb.ServiceDescriptorProto, content *strings.Builder) {

}

func generateComments(sourceCodeInfo *descriptorpb.SourceCodeInfo, content *strings.Builder) {

}

func getNativeTypeName(field *descriptorpb.FieldDescriptorProto, message *descriptorpb.DescriptorProto, pkgName string, fileExports []string) string {
	repeatedStr := ""
	if field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		repeatedStr = "[]"
	}
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
		return "number" + repeatedStr
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "boolean" + repeatedStr
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string" + repeatedStr
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "Uint8Array" + repeatedStr
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		// TODO: this lookup is not efficient, but it'll do for now. building a map of known types by name as we go would be good
		for _, nestedMessage := range message.GetNestedType() {
			// FIXME: it is possible for this to misfire at least sometimes, though we'll see if it particularly matters
			if nestedMessage.GetOptions().GetMapEntry() && strings.Contains(field.GetTypeName(), nestedMessage.GetName()) {
				keyType := getNativeTypeName(nestedMessage.GetField()[0], nil, pkgName, fileExports)
				valType := getNativeTypeName(nestedMessage.GetField()[1], nil, pkgName, fileExports)
				return fmt.Sprintf("ReadonlyMap<%s, %s | null>", keyType, valType)
			}
		}
		fieldTypeName := strings.TrimLeft(field.GetTypeName(), ".")
		if len(fieldTypeName) >= len(googleProtobufPrefix) && fieldTypeName[:len(googleProtobufPrefix)] == googleProtobufPrefix {
			// This is a google well-known type
			return fieldTypeName
		}
		// Not a map, not a google type
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		typeName := field.GetTypeName()
		matches := packageReplacement.FindStringSubmatch(typeName)
		if len(matches) != 3 {
			panic(fmt.Errorf("type name did not match any valid pattern: %s, found %d instead of 3: %s", typeName, len(matches), matches))
		}
		pkgSection := fmt.Sprintf("%s__", matches[1])
		typeSection := strings.ReplaceAll(matches[2], ".", "__") + repeatedStr
		for _, exp := range fileExports {
			if exp == typeSection {
				return typeSection
			}
		}
		return fmt.Sprintf("%s%s", pkgSection, typeSection)
	default:
		panic(fmt.Errorf("unknown field type: %s", field))
	}
}

func getProtoJSONTypeName(field *descriptorpb.FieldDescriptorProto, nestedTypes []*descriptorpb.DescriptorProto) string {
	panic("not implemented")
}
