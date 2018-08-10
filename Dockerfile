FROM alpine:3.8

RUN \
  apk add -U openssh-client git && \
  mkdir -p $HOME/.ssh && \
  ssh-keyscan github.com >> $HOME/.ssh/known_hosts && \
  chmod 0700 $HOME/.ssh && \
  chmod 0600 $HOME/.ssh/known_hosts

COPY github-bot /usr/local/bin/github-bot

CMD ["/usr/local/bin/github-bot", "-logtostderr", "-v=6"]
