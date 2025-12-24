# SUBDIRS := data mesh messanger otto server station utils
PIENV	= env GOOS=linux GOARCH=arm GOARM=7
BINARY_NAME=otto
VERSION?=0.0.12	

all: test build

init:
	git update --init 

fmt:
	gofmt -s -w .

vet:
	go vet ./...

build:
	go build -o ${BINARY_NAME} -ldflags "-X github.com/rustyeddy/otto/cmd.version=${VERSION}" ./cmd/otto

run: build
	./otto

test:
	rm -f cover.out
	go test -coverprofile=cover.out -cover ./...

verbose:
	rm -f cover.out
	go test -v -coverprofile=cover.out -cover ./...

coverage: test
	go tool cover -func=cover.out

html: test
	rm -f coverage.html
	go tool cover -html=cover.out -o coverage.html

clean:
	rm -f ${BINARY_NAME}
	rm -f cover.out coverage.html

ci: fmt vet test build

# Installation targets
INSTALL_DIR=/opt/otto
BIN_DIR=$(INSTALL_DIR)/bin
DATA_DIR=$(INSTALL_DIR)/data
SERVICE_FILE=otto.service
SYSTEMD_DIR=/etc/systemd/system

install: build
	@echo "Installing otto to $(BIN_DIR)..."
	sudo mkdir -p $(BIN_DIR)
	sudo mkdir -p $(DATA_DIR)
	sudo cp $(BINARY_NAME) $(BIN_DIR)/$(BINARY_NAME)
	sudo chmod +x $(BIN_DIR)/$(BINARY_NAME)
	@echo "Creating otto user if not exists..."
	@id -u otto &>/dev/null || sudo useradd -r -s /bin/false -d $(INSTALL_DIR) otto
	sudo chown -R otto:otto $(INSTALL_DIR)
	@echo "Installation complete: $(BIN_DIR)/$(BINARY_NAME)"

install-service: install
	@echo "Installing systemd service..."
	sudo cp $(SERVICE_FILE) $(SYSTEMD_DIR)/$(SERVICE_FILE)
	sudo systemctl daemon-reload
	@echo "Service installed. Enable with: sudo systemctl enable otto.service"
	@echo "Start with: sudo systemctl start otto.service"

enable-service:
	@echo "Enabling and starting otto service..."
	sudo systemctl enable $(SERVICE_FILE)
	sudo systemctl start $(SERVICE_FILE)
	sudo systemctl status $(SERVICE_FILE)

uninstall-service:
	@echo "Stopping and disabling otto service..."
	-sudo systemctl stop $(SERVICE_FILE)
	-sudo systemctl disable $(SERVICE_FILE)
	sudo rm -f $(SYSTEMD_DIR)/$(SERVICE_FILE)
	sudo systemctl daemon-reload
	@echo "Service uninstalled"

uninstall: uninstall-service
	@echo "Removing otto installation..."
	sudo rm -rf $(INSTALL_DIR)
	@echo "Uninstallation complete"

service-status:
	sudo systemctl status $(SERVICE_FILE)

service-logs:
	sudo journalctl -u $(SERVICE_FILE) -f

.PHONY: all build clean ci fmt run test vet install install-service enable-service uninstall-service uninstall service-status service-logs $(SUBDIRS)
