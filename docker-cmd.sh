#!/bin/bash

eval $(ssh-agent -s)

ssh-add ${GIT_SSH_KEY}

"$@"