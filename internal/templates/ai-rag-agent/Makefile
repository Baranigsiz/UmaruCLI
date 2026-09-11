.PHONY: install dev docker-up docker-down clean

install:
	pip install -r requirements.txt

dev:
	uvicorn app.main:app --reload --port 8000

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

clean:
	rm -rf __pycache__ .pytest_cache
