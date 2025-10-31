.PHONY: build run dev docker-build docker-up docker-down logs test clean

build:
	go build -o bin/compressor-api cmd/api/main.go

run:
	go run cmd/api/main.go

dev:
	air -c .air.toml

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

logs:
	docker-compose logs -f app

test:
	go test -v ./...

clean:
	rm -rf bin/
	rm -rf tmp/
	docker-compose down -v

deploy:
	docker-compose -f docker-compose.yml up -d --build

restart:
	docker-compose restart app

status:
	docker-compose ps
