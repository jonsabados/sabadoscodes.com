.DEFAULT_GOAL := dist/index.html

.env.local:
	./gen_env.sh

dist/index.html: $(shell find src) $(shell find public) .env.local
	npm run build

.PHONY: test
test:
	npm run test:unit

.PHONY: clean
clean:
	rm -rf dist/

.PHONY: run
run: .env.local
	npm run serve