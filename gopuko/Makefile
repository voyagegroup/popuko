DIST_NAME := popuko
CONFIGURE_FILE := ./config.go

all: build

clean:
	rm -rf ./$(DIST_NAME)

build: $(DIST_NAME)

build_linux_x64:
	env GOOS=linux GOARCH=amd64 make build -C .

run: $(DIST_NAME)
	./$(DIST_NAME)

$(CONFIGURE_FILE):
	cp $@.example $@

$(DIST_NAME): clean $(CONFIGURE_FILE)
	go build -o $(DIST_NAME)
