migrate-up:
	cd backend/migrations && goose up

migrate-down:
	cd backend/migrations && goose down

build-front:
	cd frontend && npm run build && rm -rf ../backend/static/* && mv dist/* ../backend/static/

run-front:
	cd frontend && npm run dev

run:
	cd backend && go run cmd/server/main.go
