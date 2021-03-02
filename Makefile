run:
	[ -e .env.dev.sh ] && source .env.dev.sh || echo "No .env.dev.sh file"
	go run .

test:
	go test -v -race

test_integration:
	[ -e .env.test.sh ] && source .env.test.sh || echo "No .env.test.sh file"
	go test -v -race -tags=integration

lint:
	golangci-lint run

build:
	go build

deploy:
	git push heroku main
