export GOPATH=$(shell pwd)/vendor

all: deps
	go build

deps: .ipfs-src
	$(MAKE) -C vendor/src/github.com/ipfs/go-ipfs deps

vendor/src/github.com/ipfs:
	mkdir -p vendor/src/github.com/ipfs

.ipfs-src: vendor/src/github.com/ipfs 
	git clone https://github.com/ipfs/go-ipfs -b master --depth=1 vendor/src/github.com/ipfs/go-ipfs
	touch .ipfs-src
