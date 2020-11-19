module gitlab.com/sapienzastudents/antispam-telegram-bot

go 1.12

replace gopkg.in/tucnak/telebot.v2 v2.3.5 => github.com/Enrico204/telebot v0.0.0-20201115170532-3dbe92edf98a

require (
	github.com/bwmarrin/discordgo v0.22.0
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/levigross/grequests v0.0.0-20190130132859-37c80f76a0da
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	gopkg.in/tucnak/telebot.v2 v2.3.5
)
