build: 
	go generate ./...
	staticcheck ./cmd/...
	CGO_ENABLED=0 go build -o main ./cmd/app/

run: build
	./main
	
watch:
	find content cmd/app/main.go cmd/app/public/images cmd/app/public/js cmd/app/public/style-md.css  cmd/gen/main.go | entr -r make run

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
