run:
	go run .

test:
	go test -v -race

test_integration:
	go test -v -race -tags=integration

lint:
	golangci-lint run

install:
	go install

deploy:
	git push heroku main
