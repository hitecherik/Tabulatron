GO = go
BINDIR = /usr/local/bin
ALL = resolver roundrunner dummyresolver tabbycatrounds zoomregistrants
LIBRARIES = $(shell find internal pkg -type f -iname '*.go')

all: $(ALL)

install: all
	cp $(ALL) /usr/local/bin

uninstall:
	cd /usr/local/bin &&\
	$(RM) $(ALL)

clean:
	$(RM) $(ALL) $(TOOLS)

resolver: cmd/resolver/resolver.go $(LIBRARIES)
	$(GO) build -o $@ $<

roundrunner: cmd/roundrunner/roundrunner.go $(LIBRARIES)
	$(GO) build -o $@ $<

dummyresolver: tools/dummyresolver/dummyresolver.go $(LIBRARIES)
	$(GO) build -o $@ $<

tabbycatrounds: tools/tabbycatrounds/tabbycatrounds.go $(LIBRARIES)
	$(GO) build -o $@ $<

zoomregistrants: tools/zoomregistrants/zoomregistrants.go $(LIBRARIES)
	$(GO) build -o $@ $<

.PHONY: all install uninstall clean
