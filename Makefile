# variables
BUILD-DIR=bin
EXE-NAME=trainer-appointment-api
SOURCE-DIR=src

build:
	cd ./$(SOURCE-DIR); go build -o "../$(BUILD-DIR)/$(EXE-NAME)"

buildRun: build run

run:
	./$(BUILD-DIR)/$(EXE-NAME)

test:
	cd ./$(SOURCE-DIR); go test
