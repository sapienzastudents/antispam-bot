package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func onAnyMessage(m *tb.Message) {
	// Note: this will not scale very well - keep an eye on it
	if !m.Private() {
		if b, err := botdb.IsBotEnabled(m.Chat); !b || err != nil {
			logger.Debugf("Bot not enabled for chat %d %s", m.Chat.ID, m.Chat.Title)
			return
		}

		botdb.UpdateMyChatroomList(m.Chat)

		textvalues := []string{
			m.Text,
			m.Caption,
			m.Payload,
		}
		if m.Photo != nil {
			textvalues = append(textvalues, m.Photo.Caption)
		}
		if m.Audio != nil {
			textvalues = append(textvalues, m.Audio.Caption)
		}
		if m.Document != nil {
			textvalues = append(textvalues, m.Document.Caption)
		}
		if m.Video != nil {
			textvalues = append(textvalues, m.Video.Caption)
		}

		for _, text := range textvalues {
			chinesePercent := chineseChars(text)
			arabicPercent := arabicChars(text)
			logger.Infof("SPAM detection (msg id %d): chinese %f arabic %f", m.ID, chinesePercent, arabicPercent)
			// Launch spam detection algorithms
			if chinesePercent > 0.05 || arabicPercent > 0.05 {
				actionDelete(m)
				// Or we can mute it (TODO: leave it as an option)
				//muteUser(m.Chat, m.Sender, m)

				break
			}
		}
	}
}
