GO = go
BINDIR = /usr/local/bin
ALL = roundrunner roundmessenger pulltabbycat tabulatron tabbycatrounds
LIBRARIES = $(shell find internal pkg -type f -iname '*.go')

all: $(ALL)

install: all
	cp $(ALL) /usr/local/bin

uninstall:
	cd /usr/local/bin &&\
	$(RM) $(ALL)

clean:
	$(RM) $(ALL) $(TOOLS)

roundrunner: cmd/roundrunner/roundrunner.go $(LIBRARIES)
	$(GO) build -o $@ $<

roundmessenger: cmd/roundmessenger/roundmessenger.go $(LIBRARIES)
	$(GO) build -o $@ $<

pulltabbycat: cmd/pulltabbycat/pulltabbycat.go $(LIBRARIES)
	$(GO) build -o $@ $<

tabulatron: cmd/tabulatron/tabulatron.go $(LIBRARIES)
	$(GO) build -o $@ $<

tabbycatrounds: tools/tabbycatrounds/tabbycatrounds.go $(LIBRARIES)
	$(GO) build -o $@ $<

.PHONY: all install uninstall clean
