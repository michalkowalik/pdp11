GO=go
MODULES = pdp/psw pdp/system pdp/unibus

build:
	$(GO) build

clean:
	$(GO) clean

tests:
	$(GO) test $(MODULES)