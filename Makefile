PLATFORM=$(shell uname | tr '[A-Z]' '[a-z]')

all: clean build-cli build-api

clean:
	@rm -f ./bin/ssm-param-api
	@rm -f ./bin/ssm-param-cli

build-api:
	@echo ayy, building ssm-param-api version $(VERSION)
	@env CGO_ENABLED=0 GOOS=$(PLATFORM) go build -a -tags netgo -ldflags '-w' -o ./bin/ssm-param-api cmd/ssm-param-api/main.go

build-cli:
	@echo ayy, building ssm-param-cli version $(VERSION)
	@env CGO_ENABLED=0 GOOS=$(PLATFORM) go build -a -tags netgo -ldflags '-w' -o ./bin/ssm-param-cli cmd/ssm-param-cli/main.go

.PHONY: all clean build-cli
