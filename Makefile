BINARY  := lume
EXTDIR  := $(HOME)/.config/timewarrior/extensions

.PHONY: build install uninstall clean

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(EXTDIR)
	ln -sf $(CURDIR)/$(BINARY) $(EXTDIR)/$(BINARY)

uninstall:
	rm -f $(EXTDIR)/$(BINARY)

clean:
	rm -f $(BINARY)
