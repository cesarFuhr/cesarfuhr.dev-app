run:
	go run main.go	

build:
	env CGO_ENABLED=0 go build -o main

docker-run:
	docker run \
		--name cesarfuhr.dev-local \
		-p 8080:8080 \
		cesarfuhr.dev/local:latest 

docker-build:
	docker build \
		-f Dockerfile.local \
		--tag cesarfuhr.dev/local \
		.

docker-clean:
	docker rm cesarfuhr.dev-local
