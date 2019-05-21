
.PHONY: build
build:
	bazel build //...

.PHONY: gazelle
gazelle:
	bazel run //:gazelle

.PHONY: deps
deps:
	go mod tidy
	go mod vendor
	bazel run //:gazelle -- update-repos -from_file=go.mod
	@make gazelle

.PHONY: run-app
run-app:
	bazel run //cmd/app:app


.PHONY: test
test:
	bazel test //...

.PHONY: clean
clean:
	bazel clean --expunge

.PHONY: push
push:
	bazel run //cmd/app:push
