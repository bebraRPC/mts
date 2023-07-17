gate:
	go run cmd/gate/main.go

worker:
	go run cmd/worker/main.go

up:
	docker-compose up

down:
	docker-compose down

rm:
	docker image rm mts-gate
	docker image rm mts-worker