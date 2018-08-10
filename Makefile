export CGO_ENABLED := 0
export IMAGE_TAG := docker.io/alvaroaleman/github-bot:3

github-bot: $(shell find . -name '*.go')
	go build -v \
		-ldflags '-s -w' \
		-o github-bot ./cmd

docker-image: github-bot
	docker build -t $(IMAGE_TAG) .
	docker push $(IMAGE_TAG)
