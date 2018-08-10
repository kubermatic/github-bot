FROM alpine:3.8

RUN \
  apk add -U openssh-client git && \
  adduser -S github-bot && \
  mkdir -p /home/github-bot/.ssh && \
  ssh-keyscan github.com >> /home/github-bot/.ssh/known_hosts && \
  chown -R github-bot /home/github-bot && \
  chmod 0700 /home/github-bot/.ssh && \
  chmod 0600 /home/github-bot/.ssh/known_hosts

USER github-bot

COPY github-bot /usr/local/bin/github-bot
