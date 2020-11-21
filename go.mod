module gitlab.com/sapienzastudents/antispam-telegram-bot

go 1.12

replace gopkg.in/tucnak/telebot.v2 v2.3.5 => github.com/Enrico204/telebot v0.0.0-20201115170532-3dbe92edf98a

require (
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/levigross/grequests v0.0.0-20190130132859-37c80f76a0da
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.3 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.7.0
	gopkg.in/tucnak/telebot.v2 v2.3.5
)
