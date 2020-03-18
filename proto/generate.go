package baetyl

//go:generate protoc -I. --go_out=plugins=grpc:. function.proto
//go:generate python3 -m grpc_tools.protoc -I. --python_out=../python36 --grpc_python_out=../python36 function.proto
//go:generate protoc-gen-grpc -I=. --js_out=import_style=commonjs,binary:../node10 --grpc_out=../node10 function.proto
