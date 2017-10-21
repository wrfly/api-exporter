.POHNY: build dev

build:
	go build .

dev: build
	./api-exporter