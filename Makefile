.PHONY: run install deploy

run:
	go run .

install:
	go install

deploy:
	git push heroku main
