package codegen

import (
	"fmt"
	"strings"

	"github.com/LLKennedy/protoc-gen-tsjson/internal/version"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var protocVersion = "unknown"

// At time of writing, the only feature that can be marked as supported is restoring the "optional" keyword to proto3, still an experimental feature that is out of scope for this project.
var support uint64 = uint64(pluginpb.CodeGeneratorResponse_FEATURE_NONE)

// Run performs code generation on the input data
func Run(request *pluginpb.CodeGeneratorRequest) (response *pluginpb.CodeGeneratorResponse) {
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

// Naive approach to codegen, creates output files for every message/service in every linked file, not just the parts depended on by the "to generate" files
func generateAllFiles(request *pluginpb.CodeGeneratorRequest) (outfiles []*pluginpb.CodeGeneratorResponse_File, err error) {
	var out *pluginpb.CodeGeneratorResponse_File
	for _, file := range request.GetProtoFile() {
		out, err = generateFullFile(file)
		if err != nil {
			return
		}
		outfiles = append(outfiles, out)
	}
	return
}

func generateFullFile(f *descriptorpb.FileDescriptorProto) (out *pluginpb.CodeGeneratorResponse_File, err error) {
	parsedName := filenameFromProto(f.GetName())
	out = &pluginpb.CodeGeneratorResponse_File{
		Name: proto.String(parsedName.fullWithoutExtension + ".ts"),
	}
	content := &strings.Builder{}
	content.WriteString(getCodeGenmarker(version.GetVersionString(), protocVersion, f.GetName()))
	generateEnums(f.GetEnumType(), content)
	out.Content = proto.String(content.String())
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
			if value.GetNumber() == 0 {
				// Special case for 0, as it doesn't get written by protojson since it's the default value
				content.WriteString(fmt.Sprintf("	/** %s */\n	%s = \"\",\n", comment, value.GetName()))
			} else {
				content.WriteString(fmt.Sprintf("	/** %s */\n	%s = \"%s\",\n", comment, value.GetName(), value.GetName()))
			}
		}
		content.WriteString("}\n\n")
	}
}
