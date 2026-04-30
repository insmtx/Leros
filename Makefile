PROJECT := singeros
REGISTRY ?= registry.yygu.cn/insmtx/

.PHONY: docker-build-singer docker-push docker-release docker-run

docker-build:
	docker build -t $(REGISTRY)$(PROJECT)-singer:latest -f deployments/build/Dockerfile.singer .

docker-push: docker-build
	docker push $(REGISTRY)$(PROJECT)-singer:latest

docker-run-singer:
	-docker rm -f $(PROJECT)-singer-dev
	docker run -d --name $(PROJECT)-singer-dev -p 8080:8080 $(REGISTRY)$(PROJECT)-singer:latest

docker-compose-up: docker-build
	docker tag $(REGISTRY)$(PROJECT)-singer:latest localhost/env_singer:latest
	docker-compose -f deployments/env/docker-compose.yml up -d

docker-compose-down:
	docker-compose -f deployments/env/docker-compose.yml down

.PHONY: run run-foreground run-detached run-build run-foreground-build run-detached-build stop logs

# Default run command - runs docker-compose services in foreground mode (shows logs)
run:
	docker-compose -f deployments/env/docker-compose.yml up

# Alternative for explicit foreground mode
run-foreground:
	docker-compose -f deployments/env/docker-compose.yml up

# Run services in foreground with forced rebuild 
run-build:
	docker-compose -f deployments/env/docker-compose.yml up --build

# Alternative for explicit foreground with forced rebuild
run-foreground-build:
	docker-compose -f deployments/env/docker-compose.yml up --build

# Run services in detached mode (background)
run-detached:
	docker-compose -f deployments/env/docker-compose.yml up -d

# Run services in detached mode with forced build
run-detached-build:
	docker-compose -f deployments/env/docker-compose.yml up -d --build

# Stop services  
stop:
	docker-compose -f deployments/env/docker-compose.yml down

# View service logs
logs:
	docker-compose -f deployments/env/docker-compose.yml logs -f

# Swagger 文档生成
.PHONY: swagger swagger-clean

swagger:
	swag init --parseDependency --generalInfo backend/cmd/singer/server.go --output docs/swagger --exclude example
	sed -i '/LeftDelim/d; /RightDelim/d' docs/swagger/docs.go

swagger-clean:
	rm -rf docs/swagger
