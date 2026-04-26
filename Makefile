BINARY  := lume
EXTDIR  := $(HOME)/.config/timewarrior/extensions

.PHONY: build install install-release uninstall clean

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(EXTDIR)
	mv $(CURDIR)/$(BINARY) $(EXTDIR)/$(BINARY)

install-release:
	./install.sh

uninstall:
	rm -f $(EXTDIR)/$(BINARY)

clean:
	rm -f $(BINARY)
