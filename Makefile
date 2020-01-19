MAIN_PKG:=az-fin
MAIN_PREFIX=$(dir $(MAIN_PKG))
MAIN=$(subst $(MAIN_PREFIX), , $(MAIN_PKG))
BIN=$(strip $(MAIN))

export GOPATH=$(shell pwd)/../../../../
export AZBIT_KUBERNETES_IDC=suzhou
export GITTAG=$(shell git describe --tags `git rev-list --tags --max-count=1`)
export GITHASH=$(shell git rev-list HEAD -n 1 | cut -c 1-)
export GITBRANCH=$(shell git symbolic-ref --short -q HEAD)

build:
	go build -tags=jsoniter -x -o run/$(BIN) . 

dev:
	go run main.go $(ARG)

run: build
	cd run && ./$(BIN) $(ARG)

init:
	cd run && TARGET='run' ARG='init' docker-compose run --rm az-fin-devel

docker-build:
	docker build . -t zzsure/az-fin:$(GITTAG) && \
	docker push zzsure/az-fin:$(GITTAG)

.PHONY: build
