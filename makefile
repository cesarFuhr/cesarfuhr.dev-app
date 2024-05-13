build: check pre
	CGO_ENABLED=0 go build -o main ./cmd/blog/

run: build
	./main

pre:
	go generate ./...

check:
	staticcheck ./cmd/...

watch:
	find 	content \
				cmd/blog/main.go \
				cmd/blog/public/images \
				cmd/blog/public/js \
				cmd/blog/public/css \
				cmd/gen | \
				entr -r make run

docker-run: docker-build
	docker run \
		--name cesarfuhr.dev \
		-p 8080:8080 \
		blog:latest 

docker-build:
	nix build '.#container'
	docker load < ./result

docker-clean:
	docker stop cesarfuhr.dev
	docker rm cesarfuhr.dev
	docker rmi blog
