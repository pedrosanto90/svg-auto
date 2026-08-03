BINARY := svg-auto
PREFIX ?= /usr/local

.PHONY: all build install uninstall clean

all: build

build:
	go build -o $(BINARY) .

install: build
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 0755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY)

clean:
	rm -f $(BINARY)
