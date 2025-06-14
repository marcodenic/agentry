.PHONY: test build serve dev

test:
	go test ./...
	cd ts-sdk && npm install && npm test

build:
	go install ./cmd/agentry

serve: build
       agentry serve --config examples/.agentry.yaml

dev: test build
       agentry serve --config examples/.agentry.yaml
