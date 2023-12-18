build:
	CGO_ENABLED=0 go build -o main

run:
	staticcheck .
	go run main.go	

watch:
	find . | entr -r make run

docker-run: docker-build
	docker run \
		--name cesarfuhr.dev \
		-p 8080:8080 \
		cesarfuhr.dev:latest 

docker-build:
	docker build \
		-f Dockerfile \
		--tag cesarfuhr.dev \
 		.

docker-clean:
	docker stop cesarfuhr.dev
	docker rm cesarfuhr.dev
	docker rmi cesarfuhr.dev
