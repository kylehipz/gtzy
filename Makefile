.PHONY: build dev-web dev-server test clean

build:
	npm --prefix web install
	npm --prefix web run build
	go build -o gtzy .

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
