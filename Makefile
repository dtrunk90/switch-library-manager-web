all: clean docker

build:
	GOOS=linux CGO_ENABLED=0 go build -o build/switch-library-manager-web main.go

clean:
	rm -rf build || true

docker: gulp build
	docker build --pull -t ghcr.io/dtrunk90/switch-library-manager-web .

gulp:
	gulp

run:
	go run main.go

watch:
	gulp watch

.PHONY: build
