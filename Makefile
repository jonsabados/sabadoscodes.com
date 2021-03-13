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

dist/authorizer: dist/ $(shell find backend/src/go)
	cd backend/src/go && GOOS=linux go build -o ../../../dist/authorizer github.com/jonsabados/sabadoscodes.com/auth/authorizer

dist/authorizerLambda.zip: dist/authorizer
	cd dist && zip authorizerLambda.zip authorizer

dist/self: dist/ $(shell find backend/src/go)
	cd backend/src/go && GOOS=linux go build -o ../../../dist/self github.com/jonsabados/sabadoscodes.com/self/lambda

dist/selfLambda.zip: dist/self
	cd dist && zip selfLambda.zip self

dist/articleAssetList: dist/ $(shell find backend/src/go)
	cd backend/src/go && GOOS=linux go build -o ../../../dist/articleAssetList github.com/jonsabados/sabadoscodes.com/article/assets/list

dist/articleAssetList.zip: dist/articleAssetList
	cd dist && zip articleAssetListLambda.zip articleAssetList

dist/articleAssetUpload: dist/ $(shell find backend/src/go)
	cd backend/src/go && GOOS=linux go build -o ../../../dist/articleAssetUpload github.com/jonsabados/sabadoscodes.com/article/assets/upload

dist/articleAssetUploadLambda.zip: dist/articleAssetUpload
	cd dist && zip articleAssetUploadLambda.zip articleAssetUpload

dist/articleSave: dist/ $(shell find backend/src/go)
	cd backend/src/go && GOOS=linux go build -o ../../../dist/articleSave github.com/jonsabados/sabadoscodes.com/article/save

dist/articleSaveLambda.zip: dist/articleSave
	cd dist && zip articleSaveLambda.zip articleSave

dist/articleGet: dist/ $(shell find backend/src/go)
	cd backend/src/go && GOOS=linux go build -o ../../../dist/articleGet github.com/jonsabados/sabadoscodes.com/article/get

dist/articleGetLambda.zip: dist/articleGet
	cd dist && zip articleGetLambda.zip articleGet

dist/articleList: dist/ $(shell find backend/src/go)
	cd backend/src/go && GOOS=linux go build -o ../../../dist/articleList github.com/jonsabados/sabadoscodes.com/article/list

dist/articleListLambda.zip: dist/articleList
	cd dist && zip articleListLambda.zip articleList

dist/backup: dist/ $(shell find backend/src/go)
	cd backend/src/go && GOOS=linux go build -o ../../../dist/backup github.com/jonsabados/sabadoscodes.com/backup/lambda

dist/backupLambda.zip: dist/backup
	cd dist && zip backupLambda.zip backup

frontend/.env.local:
	cd frontend && ./gen_env.sh

frontend/dist/index.html: $(shell find frontend/src) $(shell find frontend/public) frontend/.env.local
	cd frontend && npm run build

.PHONY: test
test:
	cd frontend && npm run test:unit
	cd backend/src/go && go test ./... --race --cover

.PHONY: clean
clean:
	rm -rf frontend/dist/ frontend/.env.local
	rm -rf dist

.PHONY: run
run: frontend/.env.local
	cd frontend && npm run serve

build: frontend/dist/index.html dist/forwarderLambda.zip dist/corsLambda.zip dist/authorizerLambda.zip dist/selfLambda.zip \
	dist/articleAssetUploadLambda.zip dist/articleAssetList.zip dist/backupLambda.zip \
	dist/articleListLambda.zip dist/articleSaveLambda.zip dist/articleGetLambda.zip