.PHONY: install run dev test docker-up docker-down

install:
	pip install -r requirements.txt

dev:
	uvicorn app.main:app --reload --port 8000

run:
	uvicorn app.main:app --host 0.0.0.0 --port 8000

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down
