# Deployment Guide

This guide explains how to deploy the Fiber Web Framework application using Docker Compose.

## Prerequisites

- Docker
- Docker Compose
- Git

## Directory Structure

```
deploy/
├── docker-compose.yml    # Docker Compose configuration
├── Dockerfile           # Docker build instructions
├── mysql/              # MySQL initialization
│   └── init.sql       # Database initialization script
└── README.md          # This file
```

## Configuration

1. Environment Variables

The following environment variables can be configured in `docker-compose.yml`:

```yaml
# Application
APP_ENV=production

# MySQL
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_USER=fiber_web
MYSQL_PASSWORD=fiber_web_password
MYSQL_DATABASE=fiber_web

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# NSQ
NSQ_LOOKUPD_HOST=nsqlookupd
NSQ_LOOKUPD_PORT=4161
```

2. Volumes

The following persistent volumes are created:
- `mysql_data`: MySQL data
- `redis_data`: Redis data

## Deployment Steps

1. Clone the repository:
```bash
git clone <repository-url>
cd fiber-web
```

2. Build and start the services:
```bash
cd deploy
docker-compose up -d
```

3. Check the status:
```bash
docker-compose ps
```

4. View logs:
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f app
```

## Service URLs

After deployment, the services will be available at:

- Application: http://localhost:3000
- MySQL: localhost:3306
- Redis: localhost:6379
- NSQ Admin UI: http://localhost:4171
- NSQ TCP: localhost:4150
- NSQ HTTP: localhost:4151

## Scaling

To scale the application:
```bash
docker-compose up -d --scale app=3
```

## Monitoring

1. Container Status:
```bash
docker-compose ps
```

2. Resource Usage:
```bash
docker stats
```

## Backup

1. MySQL Backup:
```bash
docker exec fiber_web_mysql mysqldump -u fiber_web -p fiber_web > backup.sql
```

2. Redis Backup:
```bash
docker exec fiber_web_redis redis-cli SAVE
```

## Troubleshooting

1. Check service logs:
```bash
docker-compose logs -f [service_name]
```

2. Restart a service:
```bash
docker-compose restart [service_name]
```

3. Rebuild a service:
```bash
docker-compose up -d --build [service_name]
```

4. Clean up:
```bash
# Stop all services
docker-compose down

# Remove volumes
docker-compose down -v
```

## Security Considerations

1. Change default passwords in production
2. Use secure network configurations
3. Regularly update dependencies
4. Enable HTTPS in production
5. Set up proper firewall rules
6. Implement rate limiting
7. Use secrets management

## Maintenance

1. Regular Updates:
```bash
# Pull latest images
docker-compose pull

# Rebuild and restart services
docker-compose up -d --build
```

2. Cleanup:
```bash
# Remove unused images
docker image prune

# Remove unused volumes
docker volume prune
```

## Health Checks

The application provides the following health check endpoints:

- `/health`: Basic health check
- `/metrics`: Prometheus metrics
- `/debug/pprof`: Performance profiling (development only)

## Support

For issues and support:
1. Check the logs: `docker-compose logs -f app`
2. Review the documentation
3. Submit an issue on GitHub
4. Contact the development team
