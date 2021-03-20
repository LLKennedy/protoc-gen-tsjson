# Build and install the latest version of the plugin
go install ..
# Build a test proto file

$Directory = "."
$IncludeRule = "*.proto"
$ExcludeRUle = [Regex]'.*google.*|.*audit/.*'
$PBPath = "./tsout"
$ProtoFiles = Get-ChildItem -path $Directory -Recurse -Include $IncludeRule | Where-Object FullName -NotMatch $ExcludeRUle
foreach ($file in $ProtoFiles) {
	protoc --proto_path="D:\Source\protoc-gen-tsjson\examples\node_modules\@llkennedy\protoc-gen-tsjson\" --proto_path="D:\Source\protoc-gen-tsjson\examples\sampleproto\core\" --proto_path="D:\Source\protoc-gen-tsjson\examples\sampleproto\external\" --proto_path="D:\Source\protoc-gen-tsjson\examples\" --tsjson_out=$PBPath $file.FullName
}