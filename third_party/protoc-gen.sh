protoc --proto_path=api/proto/v1 --proto_path=third_party --go_out=pkg/api/v1 --go_opt=paths=source_relative --go-grpc_out=pkg/api/v1 --go-grpc_opt=paths=source_relative --grpc-gateway_out=pkg/api/v1 --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true --swagger_out=logtostderr=true:api/swagger/v1 todo-service.proto