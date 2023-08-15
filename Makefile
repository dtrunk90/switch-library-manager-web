all: clean gulp build

build:
	GOOS=linux CGO_ENABLED=0 go build -o build/switch-library-manager-web main.go

clean:
	rm -rf build || true

gulp:
	gulp

run:
	go run main.go

watch:
	gulp watch

.PHONY: build
