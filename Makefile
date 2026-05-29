APP=url_shortner

.PHONY: run build clean

build:
	docker build -t $(APP) .

run: build
	docker run --rm -p 8000:8000 $(APP)

clean:
	docker rmi $(APP)
