# AntiSPAM Telegram Bot

This is a simple anti-spam/helper telegram bot designed for Sapienza groups.

Note that this bot will not have features that you can find in other bots like `GroupHelp`.

This bot is still a work in progress.

## Commands

#### In public groups

| Command | Admin only | Description |
| ----- | ----- | ----- |
| `/id` | No | Shows the current group ID and user ID |
| `/groups` | No | Send a private message to the user with the list of groups. If the user never started the bot, a message will be temporary sent to the group, citing the user, asking him/her to talk to the bot privately |
| `/dont` | No | Will send a message with a link to https://dontasktoask.com/ . To use this command you need to cite the message of the user (i.e. the same message will be cited by the bot). |
| `/settings` | Yes | If sent by an admin, shows the control panel for the group |
| `/terminate` | Yes | Will ban the user in 10 seconds. To use this command, cite a message of the user you want to ban. |
| `/reload` | Yes | Re-read the group admin list, group infos and bot permissions in the group |
| `/sigterm` | Yes | Terminate the bot (will delete all chat infos/settings, and the bot will leave the chatroom) |

#### As private message

| Command | Description |
| ----- | ----- |
| `/id` | Shows the current chat ID and user ID |
| `/start` | Replies with a tiny help message and two buttons: Groups and Settings |
| `/groups` | Replies with the list of categories |
| `/settings` | Replies with a list of groups where the user is admin. By clicking on a group, you will be presented the group settings view |

#### Global administrative commands (only bot admins)

| Command | Description |
| ----- | ----- |
| `/sighup` | Do a full groups cache update |
| `/groupscheck` | Prints a debug for all groups |
| `/version` | Print version/build |
| `/updatewww` | Update the group list in the website |
| `/gline` | Ban a user globally (for spam) |
| `/remove_gline` | Un-ban a user globally (for spam) |

#### Help text for BotFather

These commands will be visible to anyone.

```
groups - List all groups of the network
dont - Send a message to the user with a link to https://dontasktoask.com/
```


## Permissions

| Permission | Required | Why is needed |
| ----- | ----- | ----- |
| Change group info | No | Not used |
| Delete messages | Yes | Used to delete: commands, spam (if enabled) |
| Ban users | Yes | Used to ban for: CAS, spam (if enabled), bot commands like `/terminate` |
| Invite users via link | Yes | Used to generate invite links for group indexing |
| Pin messages | No | Not used |
| Manage video chats | No | Not used |
| Remain anonymous | No | DO NOT GRANT* |
| Add new admins | No | Not used |

*: Telegram APIs and `telebot` have a bug handling the "Remain anonymous" permission. If granted to a bot, the bot won't
be able to use all other permissions. Do not grant until we have a reasonable fix.

**Note: we strongly suggest you to grant all permissions to the bot**, especially if you plan to use the "admin fallback"
functionality. Rationale: when admins/bots promote a new admin, they can grant to the new admin only those permissions
that they have; so, if you use the "admin fallback" feature, and you never granted "Change group info" to the bot, then
**no admins will have this permissions, ever**.

**Note 2**: the bot was not tested with less than full permissions. If you encounter any bug, please create an issue.

## Features / ToDo

When not ticked, they're planned:

* [X] Use CAS blacklist
* [ ] Multilanguage!
* [X] Write an help/welcome message
* UI/UX
  * [ ] Better messages for users
  * [ ] Icons in buttons
  * [ ] Place buttons in a better way
* [ ] Merge "chatrooms" and "settings" sets in Redis
* [ ] Per-group configurable checks
  * [x] Detect Chinese messages (when Chinese chars are higher than a given
    threshold)
  * [X] Detect Arabic messages
  * [ ] Detect spam links with some heuristic
  * [ ] Detect people that are sending multiple messages for a single phrase
  * [ ] Auto-kick deleted accounts
* [x] Index all groups and provide invite links (opt-out by group admin)
  * [X] Split groups by course (degree program)
* [ ] Create an instance for other networks with all those services:
  * [ ] Discord
* [ ] Create a way to index answers and try to provide some clues (maybe some
  AI/NLP?)
* [X] Publish the group index on sapienzahub.it website
* [ ] Public log channel with actions (for auditing)
* [X] Group activity metrics (export to `/metrics` endpoint)
* [ ] Clean up the code
  * [ ] Do not use global variables (!!!)
  * [ ] Handle errors correctly (!!!)
* [x] Switch to `logrus` for structured logging
  * [ ] Use structured logging fields for logging
* [ ] Expose HTTP endpoint (optionally)
* [ ] Rewrite the bot using `telebot.v3` library to fix some structural issues
  (such as error handling, fields, int64, stateful buttons, etc)
* [ ] Write some documentation
* [ ] Write a "fallback system" for lost groups
  * When a group lose all admins, and the bot can promote a new admin, there
    should be a system capable of either promoting a new admin based on some
    algorithm, or do an election for a new admin
* [ ] Rename "categories" as "degree programs"
* [ ] Remove group list message after 10 minutes (as links will expire)

## How to build

If you want the production version (stripped/compressed binary), use `make antispam-telegram-bot`

To build a development version (not stripped, uncompressed), use `make dev`

To build a Docker container, use `make docker`

## How to spin your own instance

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