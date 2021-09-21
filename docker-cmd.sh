#!/bin/bash

eval $(ssh-agent -s)

ssh-add ${GIT_SSH_KEY}

unset GIT_SSH_KEY

exec "$@"