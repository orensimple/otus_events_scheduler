# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
SOURCE_NAME=./cmd/scheduler.go
BINARY_NAME=scheduler

all: deps gen build
gen:
	protoc --go_out=plugins=grpc:internal/grpc api/*.proto
deps:
	$(GOGET) go.uber.org/zap
	$(GOGET) go.uber.org/zap/zapcore
	$(GOGET) github.com/spf13/cobra
	$(GOGET) github.com/spf13/viper
	$(GOGET) github.com/stretchr/testify/assert
	$(GOGET) gopkg.in/natefinch/lumberjack.v2
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
run:
	$(GOBUILD) -o $(SOURCE_NAME) -v ./...
	./$(BINARY_NAME)  scheduler --config ./config
