BINARY_NAME=dropgit
GO_BUILD_ENV=CGO_ENABLED=1
GO_LDFLAGS=-ldflags="-linkmode external -extldflags -static"

.PHONY: all build install uninstall service-install service-remove test clean db-migrate dev update

all: build

build:
	$(GO_BUILD_ENV) go build $(GO_LDFLAGS) -o $(BINARY_NAME) ./cmd/dropgit

install: build
	sudo cp $(BINARY_NAME) /usr/local/bin/
	mkdir -p ~/.config/dropgit
	cp config.yml ~/.config/dropgit/config.yml.default

update: build
	systemctl --user stop dropgit.service || true
	sudo cp $(BINARY_NAME) /usr/local/bin/
	systemctl --user start dropgit.service

uninstall:
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	rm -rf ~/.config/dropgit
	rm -rf ~/.local/share/dropgit

service-install:
	mkdir -p ~/.config/systemd/user
	cp dropgit.service ~/.config/systemd/user/
	systemctl --user daemon-reload
	systemctl --user enable dropgit.service
	systemctl --user start dropgit.service

service-remove:
	systemctl --user stop dropgit.service || true
	systemctl --user disable dropgit.service || true
	rm -f ~/.config/systemd/user/dropgit.service
	systemctl --user daemon-reload

test:
	go test -v -cover ./...

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out

db-migrate: build
	./$(BINARY_NAME) -once

dev: build
	./$(BINARY_NAME)
