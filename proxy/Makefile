
all: server_stream_test.go client_stream_test.go

server_stream_test.go:
	mockery -name ServerStream -recursive -outpkg proxy -dir ../vendor/google.golang.org/grpc/ -print > server_stream_test.go

client_stream_test.go:
	mockery -name ClientStream -recursive -outpkg proxy -dir ../vendor/google.golang.org/grpc/ -print > client_stream_test.go

