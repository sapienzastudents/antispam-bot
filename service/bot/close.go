package bot

func (bot *telegramBot) Close() error {
	bot.telebot.Stop()
	return nil
}
