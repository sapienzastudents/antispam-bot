# AntiSPAM Telegram Bot

This is a simple anti-spam telegram bot designed for groups.

**WORK IN PROGRESS**: do not use in production envs

Features (when not ticked, they're planned):

* [ ] Configurable
* [x] Detect Chinese messages (when Chinese chars are higher than a given threshold)
* [X] Detect Arabic messages
* [ ] Detect spam links sent by users that sent no messages before
* [ ] Detect people that are sending multiple messages for a single phrase

Actions are:

* Ban
* Delete message (if applicable)

## How to build

If you want the production version (stripped/compressed binary), use `make antispam-telegram-bot`

To build a development version (not stripped, uncompressed), use `make dev`

To build a Docker container, use `make docker`

## Usage

The executable needs these environment variables set:

* `BOT_TOKEN`: the bot token that you get from `BotFather`
* `REDIS_URL`: the URL for a Redis server instance

To launch the executable, do the following:
```
$ export BOT_TOKEN=aaaaaaaaaaaaaa
$ export REDIS_URL=redis://localhost:6379/0
$ ./antispam-telegram-bot
```

However is strongly suggested to build a Docker container.
A pre-built one is hosted on Docker Hub at the address `enrico204/antispam-telegram-bot`.

## Contributions

To contribute please open a merge request. All code should be under the current license
(see `LICENSE` file).