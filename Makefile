.PHONY: compile invoke
.DEFAULT_GOAL := help
VERSION := 0.0.5
COMMIT_HASH := $(shell git rev-parse --short HEAD)
CURRENT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
DEFAULT_BRANCH := main
EXECUTABLE := cloudwatch-alert
S3_BUCKET := aaa
S3_KEY := bbb
help:           ## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

compile: ## delete/rebuild the go binary in  bin/
	@mkdir -p bin
	@rm -f bin/$(EXECUTABLE)
	docker run -e GOOS=linux -e GOARCH=amd64 \
	-v $$(pwd)/function:/function \
	-v $$(pwd)/bin:/bin \
	-w /function golang:1.15 go build -ldflags="-s -w" -o /bin/$(EXECUTABLE)
	(cd bin && zip ../$(EXECUTABLE).zip $(EXECUTABLE))

invoke: ## invoke the lambda
	aws lambda invoke \
    --function-name $(EXECUTABLE) \
    --payload file://event.json \
    out.json
