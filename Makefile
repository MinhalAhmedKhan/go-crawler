default: run

unit-tests:
	@echo "======  Unit Tests  ======"
	@go test -shuffle=on -v --tags=unit ./...

run:
	go run cmd/*.go
