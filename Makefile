APP=url_shortner

.PHONY: lint build run clean

lint:
	golangci-lint run ./...

build: lint
	docker build -t $(APP) .

run: build
	docker run --rm -v ./data:/data -p 8000:8000 $(APP)

clean:
	docker rmi $(APP)
