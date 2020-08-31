GO = go
BINDIR = /usr/local/bin
ALL = resolver roundrunner pulltabbycat tabulatron tabbycatrounds zoomregistrants
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

pulltabbycat: cmd/pulltabbycat/pulltabbycat.go $(LIBRARIES)
	$(GO) build -o $@ $<

tabulatron: cmd/tabulatron/tabulatron.go $(LIBRARIES)
	$(GO) build -o $@ $<

tabbycatrounds: tools/tabbycatrounds/tabbycatrounds.go $(LIBRARIES)
	$(GO) build -o $@ $<

zoomregistrants: tools/zoomregistrants/zoomregistrants.go $(LIBRARIES)
	$(GO) build -o $@ $<

.PHONY: all install uninstall clean
