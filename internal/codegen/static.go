package codegen

import "fmt"

const codegenmarker = `/**
 * Code generated by protoc-gen-tsjson. DO NOT EDIT.
 * versions:
 * 	protoc-gen-tsjson %s
 * 	protoc            %s
 * source: %s
 */
 `

func getCodeGenmarker(version, protocVersion, sourceFile string) string {
	return fmt.Sprintf(codegenmarker, version, protocVersion, sourceFile)
}
