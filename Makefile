PROJECT := SingerOS
REGISTRY ?= registry.yygu.cn/insmtx/

docker-build:

docker-push:
	docker push $(REGISTRY)$(PROJECT):latest

docker-release: docker-build docker-push

docker-run:
	-docker rm -f $(PROJECT)-dev
	docker run -d --name $(PROJECT)-dev -p 8080:8080 $(REGISTRY)$(PROJECT):latest
