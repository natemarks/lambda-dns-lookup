.PHONY: compile invoke
.DEFAULT_GOAL := help
VERSION := 0.1.0
COMMIT_HASH := $(shell git rev-parse --short HEAD)
CURRENT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
MAIN_BRANCH := master
EXECUTABLE := lambda-dns-lookup
S3_BUCKET := $(shell jq -r ".s3_bucket" config.json)
S3_KEY := $(shell jq -r ".s3_key" config.json)

help:           ## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'


clean-venv:
	rm -rf .venv
	python3 -m venv .venv
	( \
       source .venv/bin/activate; \
       pip install --upgrade pip setuptools; \
    )


compile: ## delete/rebuild the go binary in  bin/
	@mkdir -p bin
	@rm -f bin/$(EXECUTABLE)
	@rm -f $(EXECUTABLE).zip
	docker run -e GOOS=linux -e GOARCH=amd64 \
	-v $$(pwd):/function \
	-v $$(pwd)/bin:/bin \
	-w /function golang:1.15 go build -ldflags="-s -w" -o /bin/$(EXECUTABLE)
	(cd bin && zip ../$(EXECUTABLE).zip $(EXECUTABLE))
	aws s3 cp $(EXECUTABLE).zip s3://$(S3_BUCKET)/$(S3_KEY)

invoke: ## invoke the lambda
	aws lambda invoke \
    --function-name $(EXECUTABLE) \
    --payload file://event.json \
    out.json


bump: clean-venv  ## bump version in main branch
ifeq ($(CURRENT_BRANCH), $(MAIN_BRANCH))
	( \
	   source .venv/bin/activate; \
	   pip install bump2version; \
	   bump2version $(part); \
	)
else
	@echo "UNABLE TO BUMP - not on Main branch"
	$(info Current Branch: $(CURRENT_BRANCH), main: $(MAIN_BRANCH))
endif
