.PHONY: all
all: run

.PHONY: run
run:
	go run . --host ${host} --pass ${pass}