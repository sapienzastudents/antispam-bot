// This is an experimental section for a Discord bot interface
package main

//import (
//	"fmt"
//	"github.com/bwmarrin/discordgo"
//	"github.com/pkg/errors"
//	"os"
//)
//
//// Maybe we can refactor this code and the Telegram one to have multiple packages (i.e. one package per "platform"?)
//type DiscordBot struct {
//	configured bool
//	dg *discordgo.Session
//}
//
//// AppVersion is the app version injected by the compiler
//var AppVersion = "dev"
//
//func main() {
//	if err := run(); err != nil {
//		_, _ = fmt.Fprintln(os.Stderr, "error: ", err)
//		os.Exit(1)
//	}
//}
//
//func run() error {
//	rand.Seed(time.Now().UnixNano())
//	_ = godotenv.Load()
//	var err error
//
//	go mainHTTP()
//
//	/*discordbot, err := NewDiscordBot()
//	if err != nil {
//		panic(err)
//	}
//	err = discordbot.Test()
//	if err != nil {
//		panic(err)
//	}
//
//	time.Sleep(60 * time.Second)
//
//	discordbot.Stop()
//	return*/
//}
//
//func NewDiscordBot() (DiscordBot, error) {
//	var err error = nil
//	if os.Getenv("DISCORD_TOKEN") == "" {
//		return DiscordBot{configured: false}, err
//	}
//
//	var bot = DiscordBot{configured: true}
//	bot.dg, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
//	if err != nil {
//		return bot, errors.Wrap(err, "error creating Discord session")
//	}
//
//	bot.dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
//	bot.dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
//		// Ignore messages from bot itself
//		if m.Author.ID == s.State.User.ID {
//			return
//		}
//
//		if m.Content == "/settings" {
//			_, _ = s.ChannelMessageSend(m.ChannelID, "Pong!")
//		}
//	})
//
//	err = bot.dg.Open()
//	if err != nil {
//		return bot, errors.Wrap(err, "error opening Discord session")
//	}
//	return bot, nil
//}
//
//func (bot *DiscordBot) Test() error {
//	if !bot.configured {
//		return nil
//	}
//
//	// "guilds" == "servers" in Discord
//	guilds, err := bot.dg.UserGuilds(10, "", "")
//	if err != nil {
//		return errors.Wrap(err, "error getting Discord guilds")
//	}
//
//	for _, v := range guilds {
//		fmt.Println(v.Name)
//
//		channels, err := bot.dg.GuildChannels(v.ID)
//		if err != nil {
//			return errors.Wrap(err, "error getting Discord guild channels for " + v.Name)
//		}
//		var chanIdx = -1
//		for idx, c := range channels {
//			fmt.Println(c.Name)
//			if c.Type == discordgo.ChannelTypeGuildText {
//				chanIdx = idx
//			}
//		}
//
//		if chanIdx > -1 {
//			invite, err := bot.dg.ChannelInviteCreate(channels[chanIdx].ID, discordgo.Invite{
//				MaxAge:    0,
//				MaxUses:   0,
//				Temporary: false,
//			})
//			if err != nil {
//				return errors.Wrap(err, "error getting Discord invite for "+v.Name)
//			}
//			fmt.Print(invite)
//		}
//	}
//	return nil
//}
//
//func (bot *DiscordBot) Stop() {
//	_ = bot.dg.Close()
//}

func main() {
}
