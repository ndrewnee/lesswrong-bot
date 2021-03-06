run:
	go run .

test:
	go test -v -race ./...

test_integration:
	go test -v -race -tags=integration ./...

lint:
	golangci-lint run

build:
	go build

deploy:
	git push heroku main
