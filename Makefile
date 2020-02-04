.DEFAULT_GOAL := build

dist/:
	mkdir dist

dist/forwarder: dist/ $(shell find backend/src/go)
	cd backend/src/go && GOOS=linux go build -o ../../../dist/forwarder github.com/jonsabados/sabadoscodes.com/mail/forwarder

dist/forwarderLambda.zip: dist/forwarder
	cd dist && zip forwarder.zip forwarder

dist/cors: dist/ $(shell find backend/src/go)
	cd backend/src/go && GOOS=linux go build -o ../../../dist/cors github.com/jonsabados/sabadoscodes.com/cors/lambda

dist/corsLambda.zip: dist/cors
	cd dist && zip corsLambda.zip cors

frontend/.env.local:
	cd frontend && ./gen_env.sh

frontend/dist/index.html: $(shell find frontend/src) $(shell find frontend/public) frontend/.env.local
	cd frontend && npm run build

.PHONY: test
test:
	cd frontend && npm run test:unit
	cd backend/src/go && go test ./... --race

.PHONY: clean
clean:
	rm -rf frontend/dist/ frontend/.env.local
	rm -rf dist

.PHONY: run
run: frontend/.env.local
	cd frontend && npm run serve

build: frontend/dist/index.html dist/forwarderLambda.zip dist/corsLambda.zip