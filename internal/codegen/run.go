package codegen

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

// At time of writing, the only feature that can be marked as supported is restoring the "optional" keyword to proto3, still an experimental feature that is out of scope for this project.
var support uint64 = uint64(pluginpb.CodeGeneratorResponse_FEATURE_NONE)

// Run performs code generation on the input data
func Run(request *pluginpb.CodeGeneratorRequest) (response *pluginpb.CodeGeneratorResponse) {
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
	generatedFiles, err := generateFiles(request)
	if err != nil {
		// It didn't work, ignore any data we generated and only return the error
		response.Error = proto.String(fmt.Sprintf("failed to generate files: %v", err))
		return
	}
	// It worked, set the response data
	response.File = generatedFiles
	return
}

func generateFiles(request *pluginpb.CodeGeneratorRequest) (outfiles []*pluginpb.CodeGeneratorResponse_File, err error) {
	outfiles = append(outfiles, &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String("out/package.json"),
		Content: proto.String(`{}`),
	})
	return
}
