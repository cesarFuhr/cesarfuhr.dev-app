run:
	go run main.go	

build:
	env CGO_ENABLED=0 go build -o main
