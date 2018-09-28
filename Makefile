compose_file="docker-compose.yml"

.PHONY: clean build install start stop

DETACH_FLAG :=
ifeq ($(DETACH),1)
export DETACH_FLAG := -d
endif

clean:
	docker-compose -f $(compose_file) down --volumes
	docker image prune --force

build:
	docker-compose -f $(compose_file) build

install:
	docker-compose -f $(compose_file) up $(DETACH_FLAG)

start:
	docker-compose -f $(compose_file) start

stop:
	docker-compose -f $(compose_file) stop
