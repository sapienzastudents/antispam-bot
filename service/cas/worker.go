package cas

import "time"

func (cas *cas) worker() {
	t := time.NewTicker(1 * time.Hour)
	for !cas.workerStop {
		_, _, _ = cas.Load()

		// Here we might automatically ban newly CAS-banned users, but for now we limit the bot to
		// react when an user do some action (to avoid flooding Telegram APIs)

		//chats, err := botdb.ListMyChatrooms()
		//if err != nil {
		//	for _, chat := range chats {
		//		settings, err := botdb.GetChatSetting(chat)
		//		if err != nil {
		//			continue
		//		}
		//		if settings.BotEnabled && settings.OnBlacklistCAS.Action != ACTION_NONE {
		//			for uid := range added {
		//				performAction(nil, &tb.User{
		//					ID: uid,
		//				}, settings.OnBlacklistCAS)
		//				time.Sleep(1 * time.Second)
		//			}
		//		}
		//	}
		//}
		<-t.C
	}
}
