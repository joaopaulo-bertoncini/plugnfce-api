#!/bin/bash

# NFC-e API Development Docker Script
# Usage: ./scripts/docker-dev.sh [up|down|build|logs|shell|test]

set -e

COMPOSE_FILE="docker-compose.yml"
PROJECT_NAME="plugnfce"

case "${1:-help}" in
    "up")
        echo "🚀 Starting NFC-e development environment..."
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d
        echo "⏳ Waiting for services to be ready..."
        sleep 10
        echo "✅ Services started!"
        echo ""
        echo "🌐 Services available:"
        echo "  - API: http://localhost:8080"
        echo "  - MinIO Console: http://localhost:9001"
        echo "  - RabbitMQ Management: http://localhost:15672"
        echo "  - PostgreSQL: localhost:5432"
        echo ""
        echo "📊 View logs: ./scripts/docker-dev.sh logs"
        echo "🐚 Shell access: ./scripts/docker-dev.sh shell [api|worker|db]"
        ;;
    "down")
        echo "🛑 Stopping NFC-e development environment..."
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down
        echo "✅ Services stopped!"
        ;;
    "build")
        echo "🔨 Building Docker images..."
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME build --no-cache
        echo "✅ Images built!"
        ;;
    "rebuild")
        echo "🔄 Rebuilding and restarting services..."
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME build --no-cache
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d
        echo "✅ Services rebuilt and restarted!"
        ;;
    "logs")
        case "${2:-all}" in
            "api")
                docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f api
                ;;
            "worker")
                docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f worker
                ;;
            "db")
                docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f db
                ;;
            "rabbitmq")
                docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f rabbitmq
                ;;
            "minio")
                docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f minio
                ;;
            *)
                docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f
                ;;
        esac
        ;;
    "shell")
        case "${2:-api}" in
            "api")
                docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec api sh
                ;;
            "worker")
                docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec worker sh
                ;;
            "db")
                docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec db bash
                ;;
            *)
                echo "Usage: $0 shell [api|worker|db]"
                exit 1
                ;;
        esac
        ;;
    "test")
        echo "🧪 Running tests..."
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec api go test ./...
        ;;
    "clean")
        echo "🧹 Cleaning up Docker resources..."
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down -v --remove-orphans
        docker system prune -f
        echo "✅ Cleanup completed!"
        ;;
    "status")
        echo "📊 Service Status:"
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME ps
        ;;
    "help"|*)
        echo "🐳 NFC-e Development Docker Script"
        echo ""
        echo "Usage: $0 [command] [options]"
        echo ""
        echo "Commands:"
        echo "  up          Start all services"
        echo "  down        Stop all services"
        echo "  build       Build Docker images"
        echo "  rebuild     Rebuild and restart services"
        echo "  logs        Show logs (api|worker|db|rabbitmq|minio)"
        echo "  shell       Open shell in container (api|worker|db)"
        echo "  test        Run tests"
        echo "  clean       Clean up Docker resources"
        echo "  status      Show service status"
        echo "  help        Show this help"
        echo ""
        echo "Examples:"
        echo "  $0 up"
        echo "  $0 logs api"
        echo "  $0 shell worker"
        ;;
esac
