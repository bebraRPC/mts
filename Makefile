gate:
	go run cmd/gate/main.go

worker:
	go run cmd/worker/main.go

up:
	docker-compose up

down:
	docker-compose down

gen:
	protoc -I . -I ./google/api --go_out=. --go-grpc_out=. --grpc-gateway_out=.  --swagger_out=./swagger --swagger_opt=logtostderr=true proto/gateway.proto

rm:
	docker rm gate
	docker rm worker
	docker image rm mts-gate
	docker image rm mts-worker