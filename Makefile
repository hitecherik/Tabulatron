GO = go
BINDIR = /usr/local/bin
ALL = roundrunner roundmessenger pulltabbycat tabulatron tabbycatrounds
LIBRARIES = $(shell find internal pkg -type f -iname '*.go')

all: $(ALL)

$(ALL): %: cmd/%/main.go $(LIBRARIES)
	$(GO) build -o $@ $<

install: all
	cp $(ALL) /usr/local/bin

uninstall:
	cd /usr/local/bin &&\
	$(RM) $(ALL)

clean:
	$(RM) $(ALL)

.PHONY: all install uninstall clean
