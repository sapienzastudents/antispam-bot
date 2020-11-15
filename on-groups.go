package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"sort"
	"strings"
)

func onGroups(m *tb.Message, _ ChatSettings) {
	logger.Debugf("My chat room requested by %d (%s %s %s)", m.Sender.ID, m.Sender.Username, m.Sender.FirstName, m.Sender.LastName)

	chatrooms, err := botdb.ListMyChatrooms()
	if err != nil {
		logger.Criticalf("Error getting chatroom list: %s", err.Error())
	} else {
		var rows [][]tb.InlineButton

		sort.Slice(chatrooms, func(i, j int) bool {
			return chatrooms[i].Title < chatrooms[j].Title
		})

		msg := strings.Builder{}
		msg.WriteString("🇮🇹 Premi sul pulsante di un gruppo per ricevere il suo link di invito\n\n🇬🇧 Click on a group button to receive its invite link")
		for _, v := range chatrooms {
			settings, err := botdb.GetChatSetting(v)
			if err != nil {
				logger.Criticalf("Error getting chatroom config: %s", err.Error())
				continue
			}
			if settings.Hidden {
				continue
			}

			s := sha1.New()
			s.Write([]byte(v.Title))

			btn := tb.InlineButton{
				Unique: hex.EncodeToString(s.Sum(nil)),
				Text:   v.Title,
				Data:   fmt.Sprintf("%d", v.ID),
			}

			b.Handle(&btn, groupButtonHandler(v))
			rows = append(rows, []tb.InlineButton{btn})
		}

		_, err = b.Send(m.Sender, msg.String(), &tb.SendOptions{
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: rows,
			},
		})
		if err != nil {
			logger.Warning("can't send message to the user ", err)
		}
	}

	if !m.Private() {
		b.Send(m.Chat, "🇮🇹 Ti ho scritto in privato!\n\n🇬🇧 I sent you a direct message!", &tb.SendOptions{ReplyTo: m})
	}
}

func groupButtonHandler(v *tb.Chat) func(callback *tb.Callback) {
	return func(callback *tb.Callback) {
		err := b.Respond(callback, &tb.CallbackResponse{})
		if err != nil {
			return
		}

		v, err = b.ChatByID(fmt.Sprint(v.ID))
		if err != nil {
			logger.Warning("can't get chat info ", err)
			return
		}

		if v.Username != "" {
			_, err = b.Send(callback.Sender, "@"+v.Username)
			if err != nil {
				logger.Warning("can't send message to the user ", err)
			}
		} else {
			if v.InviteLink == "" {
				inviteLink, err := b.GetInviteLink(v)

				if err != nil && err.Error() == tb.ErrGroupMigrated.Error() {
					apierr, _ := err.(*tb.APIError)
					v, err = b.ChatByID(fmt.Sprint(apierr.Parameters["migrate_to_chat_id"]))
					if err != nil {
						logger.Warning("can't get chat info ", err)
						return
					}

					_ = botdb.UpdateMyChatroomList(v)
					inviteLink, err = b.GetInviteLink(v)
					if err != nil {
						logger.Warning("can't get invite link ", err)
						return
					}
				}

				if err != nil {
					logger.Warning("can't get the invite link for chat ", v, err)
					/*_, err = b.Send(v, "Oops, qualcuno mi ha chiesto il link di invito di questo gruppo ma non posso generarlo, mi mancano i permessi (di admin)! :-)")
					if err != nil {
						logger.Warning("can't send message to the group ", err)
					}*/

					_, err = b.Send(callback.Sender, "Oops, non riesco a prendere il link, probabilmente il gruppo non esiste più oppure non sono admin. Riprova più tardi!")

					if err != nil {
						logger.Warning("can't send message to the user ", err)
					}
					return
				} else if inviteLink == "" {
					logger.Warning("invite link empty for ", v)
					return
				}

				_, err = b.Send(callback.Sender, "Link: "+inviteLink)
				if err != nil {
					logger.Warning("can't send message to the user ", err)
				}
			} else {
				_, err = b.Send(callback.Sender, "Link: "+v.InviteLink)
				if err != nil {
					logger.Warning("can't send message to the user ", err)
				}
			}
		}
	}
}
