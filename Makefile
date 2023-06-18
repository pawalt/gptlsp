.PHONY: vendor
vendor:
	go mod vendor
	./clone_submodules.sh vendor/github.com/go-skynet/go-llama.cpp

.PHONY: preload
preload:
	cd vendor/github.com/go-skynet/go-llama.cpp && \
	  BUILD_TYPE=metal make libbinding.a
	cp vendor/github.com/go-skynet/go-llama.cpp/build/bin/ggml-metal.metal .


.PHONY: build
build:
	CGO_LDFLAGS="-framework Foundation -framework Metal -framework MetalKit -framework MetalPerformanceShaders" \
		LIBRARY_PATH=$PWD C_INCLUDE_PATH=$PWD \
		go build .

.PHONY: run
run: build
	./gptlsp