GO=go
MODULES = pdp/mmu pdp/psw pdp/system pdp/unibus

build:
	$(GO) build

clean:
	$(GO) clean

tests:
	$(GO) test $(MODULES)