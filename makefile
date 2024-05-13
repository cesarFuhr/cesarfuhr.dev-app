app: check pre
	CGO_ENABLED=0 go build -o main ./cmd/app/

run: build
	./main

pre:
	go generate ./...

check:
	staticcheck ./cmd/...

watch:
	find 	content \
				cmd/app/main.go \
				cmd/app/public/images \
				cmd/app/public/js \
				cmd/app/public/css \
				cmd/gen | \
				entr -r make run

docker-run: docker-build
	docker run \
		--name cesarfuhr.dev \
		-p 8080:8080 \
		cesarfuhr.dev:latest 

docker-build:
	make build
	docker build \
		-f Dockerfile \
		--tag cesarfuhr.dev \
 		.

docker-clean:
	docker stop cesarfuhr.dev
	docker rm cesarfuhr.dev
	docker rmi cesarfuhr.dev
