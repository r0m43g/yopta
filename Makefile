migrate-up:
	cd backend/migrations && goose up

migrate-down:
	cd backend/migrations && goose down

build-front:
	cd frontend && npm run build && rm -rf ../backend/static/* && mv dist/* ../backend/static/

build-back:
	cd backend && go build -o yopta-server cmd/server/main.go

build: build-front build-back

build-prod:
	cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o yopta-server cmd/server/main.go

run-front:
	cd frontend && npm run dev

run-back:
	cd backend && ./yopta-server

run:
	cd backend && go run cmd/server/main.go

clean:
	rm -f backend/yopta-server
	rm -rf backend/static/*
	rm -rf frontend/dist

