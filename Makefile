FRSSOURCES := $(shell find . -name "*.go")
PRUNERSOURCES := $(shell find . -name "*.go")
UNCURRENTERSOURCES := $(shell find . -name "*.go")

all: frs/frs pruner/pruner uncurrenter/uncurrenter

frs/frs: $(FRSSOURCES)
	cd frs && go build .

pruner/pruner: $(PRUNERSOURCES)
	cd pruner && go build .

uncurrenter/uncurrenter: $(UNCURRENTERSOURCES)
	cd uncurrenter && go build .
