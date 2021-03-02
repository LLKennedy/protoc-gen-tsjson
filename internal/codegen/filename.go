package codegen

import "strings"

type filename struct {
	pathParts []string
	name      string
}

func filenameFromProto(in string) (out filename) {
	parts := strings.Split(in, "/")
	lastPart := parts[len(parts)-1]
	if len(parts) > 1 {
		out.pathParts = parts[:len(parts)-1]
	}
	out.name = strings.Split(lastPart, ".proto")[0]
	return
}
