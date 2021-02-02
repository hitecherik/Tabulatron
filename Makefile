GO = go
GOFMT = gofmt -s
BINDIR = /usr/local/bin
ALL = roundrunner roundmessenger pulltabbycat tabulatron tabbycatrounds
LIBRARIES = $(shell find internal pkg -type f -iname '*.go')

all: $(ALL)

$(ALL): %: cmd/%/main.go $(LIBRARIES)
	$(GO) build -o $@ $<

fmt:
	$(GOFMT) -w $(shell find . -type f -iname '*.go')

checkfmt:
	$(GOFMT) -l $(shell find . -type f -iname '*.go') |\
	wc -l |\
	perl -e 'my $$lines = <STDIN>; chomp $$lines; exit $$lines == 0;'

install: all
	cp $(ALL) /usr/local/bin

uninstall:
	cd /usr/local/bin &&\
	$(RM) $(ALL)

clean:
	$(RM) $(ALL)

.PHONY: all install uninstall clean
