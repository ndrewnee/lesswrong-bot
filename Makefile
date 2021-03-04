run:
	# source .env.dev
	go run .

test:
	go test -v -race ./...

test_integration:
	# source .env.test
	go test -v -race -tags=integration ./...

lint:
	golangci-lint run

build:
	go build

deploy:
	git push heroku main
