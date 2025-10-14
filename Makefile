# Makefile for Messenger Docker Compose project

.PHONY: help build up down restart logs logs-app logs-postgres logs-kafka logs-zookeeper console-postgres console-app console-kafka ps clean db-shell test

# Default target
help:
	@echo "Available commands:"
	@echo "  make build        - Build or rebuild services"
	@echo "  make up          - Create and start containers"
	@echo "  make down        - Stop and remove containers"
	@echo "  make restart     - Restart all services"
	@echo "  make ps          - Show container status"
	@echo ""
	@echo "Logs:"
	@echo "  make logs        - View logs from all services"
	@echo "  make logs-app    - View app logs"
	@echo "  make logs-postgres - View PostgreSQL logs"
	@echo "  make logs-kafka  - View Kafka logs"
	@echo "  make logs-zookeeper - View Zookeeper logs"
	@echo ""
	@echo "Console access:"
	@echo "  make console-postgres - Access PostgreSQL console"
	@echo "  make console-app     - Access app container shell"
	@echo "  make console-kafka   - Access Kafka container"
	@echo "  make db-shell       - Access PostgreSQL shell"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean        - Stop and remove containers, networks, volumes"
	@echo "  make test         - Run tests"

# Build services
build:
	docker-compose build

# Start services in background
up:
	docker-compose up -d

# Build and start
up-build: build up

# Stop services
down:
	docker-compose down

# Restart services
restart:
	docker-compose restart

# Show status
ps:
	docker-compose ps

# View all logs
logs:
	docker-compose logs -f

# View specific service logs
logs-app:
	docker-compose logs -f app

logs-postgres:
	docker-compose logs -f postgres

logs-kafka:
	docker-compose logs -f kafka

logs-zookeeper:
	docker-compose logs -f zookeeper

logs-kafka-ui:
	docker-compose logs -f kafka-ui

# Console access
console-postgres:
	docker-compose exec postgres psql -U sibhelly -d db

console-app:
	docker-compose exec app sh

console-kafka:
	docker-compose exec kafka bash

# PostgreSQL shell access
db-shell:
	docker-compose exec postgres bash

# Clean everything
clean:
	docker-compose down -v --remove-orphans
	docker system prune -f

# Development shortcuts
dev: up-build logs-app

# Database operations
db-backup:
	docker-compose exec postgres pg_dump -U sibhelly db > backup_$(shell date +%Y%m%d_%H%M%S).sql

db-restore:
	@echo "Usage: docker-compose exec -T postgres psql -U sibhelly db < your_backup_file.sql"

# Test commands
test:
	@echo "Running tests..."
	# Добавьте здесь команды для тестов

# Health checks
health:
	@echo "=== Checking services health ==="
	@echo "PostgreSQL: $$(docker-compose exec postgres pg_isready -U user -d db && echo '✅' || echo '❌')"
	@echo "Kafka: $$(docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092 >/dev/null 2>&1 && echo '✅' || echo '❌')"
	@echo "App: $$(curl -s http://localhost:8080/health >/dev/null && echo '✅' || echo '❌')"

# Quick start for development
start: up-build
	@echo "Services are starting..."
	@sleep 10
	@make health

# Monitor all services
monitor:
	@echo "Monitoring all services (Ctrl+C to stop)"
	docker-compose logs -f

# Short aliases
b: build
u: up
d: down
r: restart
la: logs-app
lp: logs-postgres
lk: logs-kafka
lz: logs-zookeeper
cp: console-postgres
ca: console-app
ck: console-kafka
ds: db-shell