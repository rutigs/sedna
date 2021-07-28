.DEFAULT_GOAL := run

run:
	go run main.go -redis=0.0.0.0:6379

build:
	go build -o app .

docker:
	docker build . -t sedna:latest

docker/run:
	docker run --rm -p 3000:3000 sedna:latest

compose: compose/build compose/up

compose/build:
	docker-compose build

compose/up:
	docker-compose up
