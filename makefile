CONF ?=$(shell pwd)/configs
.PHONY: run-entrepot
generate:
	dapr run --app-id entrepot --app-port 81 --components-path ${CONF}/samples -- go run cmd/main.go --config ${CONF}/config.yml