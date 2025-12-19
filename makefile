run:
	go mod tidy && go run ./cmd/main.go

build:
	go build -o bin/app ./cmd/main.go

clean:
	rm -rf bin
