# Makefile
# Mục tiêu 'gen-tel'
gen-tel:
    protoc --go_out=. --go-grpc_out=. chat.proto