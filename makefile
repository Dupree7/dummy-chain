COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%d %H:%M:%S")

METADATA_DIR := metadata
APP_NAME_CLIENT := dummyclient
APP_NAME_VALIDATOR := dummyvalidator

.PHONY: client
client: clean metadata
	@echo "package metadata\n" > $(METADATA_DIR)/role.go
	@echo "const Role = \"client\"" >> $(METADATA_DIR)/role.go

	@echo "Building $(APP_NAME) node"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(DATE)"
	go build -o $(APP_NAME_CLIENT)

.PHONY: validator
validator: clean metadata
	@echo "package metadata\n" > $(METADATA_DIR)/role.go
	@echo "const Role = \"validator\"" >> $(METADATA_DIR)/role.go

	@echo "Building $(APP_NAME) node"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(DATE)"
	go build -o $(APP_NAME_VALIDATOR)

.PHONY: metadata
metadata:
	@echo "package metadata\n" > $(METADATA_DIR)/git_commit.go
	@echo "const GitCommit = \"$(COMMIT)\"" >> $(METADATA_DIR)/git_commit.go
	@echo "package metadata\n" > $(METADATA_DIR)/build_date.go
	@echo "const BuildDate = \"$(DATE)\"" >> $(METADATA_DIR)/build_date.go

.PHONY: clean
clean:
	rm -f $(APP_NAME_CLIENT) $(APP_NAME_VALIDATOR)