APP_EXT := $(if $(filter Windows_NT,$(OS)),.exe)

build:
	@go build -o padding-size$(APP_EXT)

install: build
	@go install


test:
	@go test
