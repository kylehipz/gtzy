.PHONY: build install web dev-web dev-server test clean

web:
	npm --prefix web install
	npm --prefix web run build

build: web
	go build -o gtzy .

install: web
	go install .

dev-web:
	npm --prefix web run dev

dev-server:
	go run . serve

test:
	go test ./...

clean:
	rm -f gtzy
	rm -rf web/dist
	mkdir -p web/dist && touch web/dist/.gitkeep
