package genproto

//go:generate protoc --go_out=. --go-grpc_out=. -I../../api ../../api/auth.proto ../../api/crud.proto ../../api/cdc.proto
