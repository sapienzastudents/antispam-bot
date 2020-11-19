# AntiSPAM Telegram Bot

This is a simple anti-spam telegram bot designed for groups.

Note that this bot will not have some features that you can find in other bots like `GroupHelp`.

**WORK IN PROGRESS**

Features (when not ticked, they're planned):

* [X] Use CAS blacklist
* [ ] Write an help/welcome message
* [ ] Merge "chatrooms" and "settings" sets in Redis
* [ ] Per-group configurable checks
    * [x] Detect Chinese messages (when Chinese chars are higher than a given threshold)
    * [X] Detect Arabic messages
    * [ ] Detect spam links with some heuristic
    * [ ] Detect people that are sending multiple messages for a single phrase
* [x] Index all groups and provide invite links (opt-out by group admin)
    * [ ] Split groups by course (degree program)
* [ ] Create an instance for other networks with all those services:
    * [ ] Discord
* [ ] Create a way to index answers and try to provide some clues (maybe some AI/NLP?)
* [X] Publish the group index on sapienzahub.it website
* [ ] Public log channel with actions (for auditing)
* [ ] Group activity metrics (export to InfluxDB/Prometheus/whatever and expose via Grafana)
* [ ] Notify on error (via Telegram? via logging?)

Also, in ToDo list:

* [ ] Clean up the code
    * [ ] Do not use global variables (!!!)
    * [ ] Handle errors correctly (!!!)
* [x] Switch to `logrus` for structured logging
    * [ ] Use structured logging fields for logging
* [ ] Expose HTTP endpoint (optionally)
* [ ] Rewrite the `telebot` library to fix some structural issues (such as error handling, fields, int64, etc)

## Commands

```
groups - List all groups of the network
settings - Group settings
terminate - Ban an user from a group with a 60s delay
```

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