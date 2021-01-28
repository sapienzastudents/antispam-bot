# AntiSPAM Telegram Bot

This is a simple anti-spam/helper telegram bot designed for Sapienza groups.

Note that this bot will not have features that you can find in other bots like `GroupHelp`.

**WORK IN PROGRESS**

Features (when not ticked, they're planned):

* [X] Use CAS blacklist
* [ ] Multilanguage!
* [X] Write an help/welcome message
* [ ] Merge "chatrooms" and "settings" sets in Redis
* [ ] Per-group configurable checks
    * [x] Detect Chinese messages (when Chinese chars are higher than a given threshold)
    * [X] Detect Arabic messages
    * [ ] Detect spam links with some heuristic
    * [ ] Detect people that are sending multiple messages for a single phrase
    * [ ] Auto-kick deleted accounts
* [x] Index all groups and provide invite links (opt-out by group admin)
    * [X] Split groups by course (degree program)
* [ ] Create an instance for other networks with all those services:
    * [ ] Discord
* [ ] Create a way to index answers and try to provide some clues (maybe some AI/NLP?)
* [X] Publish the group index on sapienzahub.it website
* [ ] Public log channel with actions (for auditing)
* [X] Group activity metrics (export to InfluxDB/Prometheus/whatever and expose via Grafana)
* [ ] Notify on error (via Telegram? via logging?)
* [ ] Restore direct telegram link in group index (web/bot): during first indexing or periodically, get the link from the chat info. If empty, generate a new link

Also, in ToDo list:

* [ ] Clean up the code
    * [ ] Do not use global variables (!!!)
    * [ ] Handle errors correctly (!!!)
* [x] Switch to `logrus` for structured logging
    * [ ] Use structured logging fields for logging
* [ ] Expose HTTP endpoint (optionally)
* [ ] Rewrite the `telebot` library to fix some structural issues (such as error handling, fields, int64, stateful buttons, etc)

## Commands

```
groups - List all groups of the network
settings - Group settings
terminate - Ban an user from a group with a 60s delay
```

## Permissions

The bot requires all admin permissions (except for "post anonymously") for the above stated features.
It requires also the "Add admins" permission an user-bot will be implemented for some features
(antispam, kick deleted accounts, etc) that are not available in classic bot APIs.

## How to build

If you want the production version (stripped/compressed binary), use `make antispam-telegram-bot`

To build a development version (not stripped, uncompressed), use `make dev`

To build a Docker container, use `make docker`

## Usage

The executable needs these environment variables set:

* `BOT_TOKEN`: the bot token that you get from `BotFather`
* `REDIS_URL`: the URL for a Redis server instance
* `DISABLE_CAS`: optional, disable CAS blacklist completely
* `GIT_TEMP_DIR`: optional, temporary directory for website update
* `GIT_SSH_KEY`: optional, Git SSH key path for website update

The executable supports `.env` file for environment variables.

A pre-built Docker container is hosted on Docker Hub at the address `enrico204/antispam-telegram-bot`.

## Contributions

To contribute please open a merge request. All code should be under the current license
(see `LICENSE` file).