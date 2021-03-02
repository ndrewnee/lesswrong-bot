run:
	go run .

test:
	go test -v -race

test_integration:
	[ -e .env.sh ] && source .env.sh || echo "No .env.sh file"
	go test -v -race -tags=integration

lint:
	golangci-lint run

install:
	go install

deploy:
	git push heroku main
