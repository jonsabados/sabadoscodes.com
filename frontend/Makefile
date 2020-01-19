.DEFAULT_GOAL := dist/index.html

dist/index.html: $(shell find src) $(shell find public)
	npm run build

.PHONY: test
test:
	npm run test:unit

.PHONY: clean
clean:
	rm -rf dist/

.PHONY: run
run:
	npm run serve