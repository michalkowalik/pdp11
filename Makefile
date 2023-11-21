GO=go
MODULES = pdp/psw pdp/system pdp/unibus

build:
	$(GO) build

clean:
	$(GO) clean

tests:
	$(GO) test $(MODULES)

tests_verbose:
	$(GO) test -v $(MODULES)

debug:
	$(GO) build -gcflags="all=-N -l"
