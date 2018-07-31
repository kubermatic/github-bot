export CGO_ENABLED := 0

cherry_pick_bot: $(shell find . -name '*.go')
	go build -v \
		-ldflags '-s -w' \
		-o cherry_pick_bot .
