package bot

import (
	"html/template"
	"strings"

	tb "gopkg.in/tucnak/telebot.v3"
)

// sendAdminsForSettings sends the bot admins settings panel to the chat where
// the message is sent.
//
// This panel can be accessed when the user clicks on admin settings button,
// inside the general settings panel.
func (bot *telegramBot) sendAdminsForSettings(sender *tb.User, msgToEdit *tb.Message) {
	lang := sender.LanguageCode

	// Only bot admins can see admins settings panel.
	if is, err := bot.db.IsBotAdmin(sender.ID); err != nil {
		bot.logger.WithError(err).Error("Failed to check if the user is a bot admin")
		return
	} else if !is {
		bot.logger.WithField("user_id", sender.ID).Warn("User is not a global admin but triggered admins settings panel")
		return
	}

	// Buttons to send with the reply message.
	var chatButtons [][]tb.InlineButton

	// Close button.
	bt := tb.InlineButton{
		Unique: "admins_settings_close",
		Text:   "🚪 " + bot.bundle.T(lang, "Close"),
	}
	bot.telebot.Handle(&bt, func(ctx tb.Context) error {
		callback := ctx.Callback()
		_ = bot.telebot.Respond(callback)
		_ = bot.telebot.Delete(callback.Message)
		return nil
	})
	chatButtons = append(chatButtons, []tb.InlineButton{bt})

	// Add new admin button.
	bt = tb.InlineButton{
		Unique: "admins_settings_add_admin",
		Text:   "➕" + bot.bundle.T(lang, "Add admin"),
	}
	bot.handleAdminCallbackStateful(&bt, bot.handleAddAdmin)
	chatButtons = append(chatButtons, []tb.InlineButton{bt})

	adminsID, err := bot.db.GetBotAdmins()
	if err != nil {
		bot.logger.WithError(err).Error("Failed to retrieve bot admins")
	}

	// GetBotAdmins returns only a slice of IDs, we need some additional
	// information to build a useful message. We do not use *tb.User because we
	// use only a small subset of fields.
	type User struct {
		ID int64

		// If FirstName, LastName and Username are all empty strings it mens the
		// user never sent a message to the bot or he blocked it. Some fields
		// can be empty.
		FirstName string
		LastName  string
		Username  string
	}

	admins := make([]User, 0, len(adminsID))
	for i, id := range adminsID {
		admins = append(admins, User{ID: id})

		chat, err := bot.telebot.ChatByID(id)
		// Maybe the bot has been blocked by the admin or he never sent a
		// message to it.
		if err != nil {
			bot.logger.WithField("user_id", id).Warn("Failed to get user's info")
			continue
		}
		admins[i].FirstName = chat.FirstName
		admins[i].LastName = chat.LastName
		admins[i].Username = chat.Username
	}

	// To build the text message we use html/template, because it generates HTML
	// output that is safe from code injection.
	adminsTpl := bot.bundle.T(lang, `
<b>Bot admins</b>

{{range $admin := . -}}
· <code>{{ $admin.ID }}</code> - {{ $admin.FirstName }} {{ $admin.LastName }} {{ if $admin.Username }} ({{ $admin.Username }}) {{ end }}
{{else}}
There are not admins, this is strange!
{{end}}
`)

	t, err := template.New("admins").Parse(adminsTpl)
	if err != nil {
		bot.logger.WithError(err).Error("Failed to parse template for admins panel")
		return
	}

	msg := &strings.Builder{}
	if err := t.Execute(msg, admins); err != nil {
		bot.logger.WithError(err).Error("Failed to execute template for admins panel")
		return
	}

	options := &tb.SendOptions{
		ParseMode:   tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{InlineKeyboard: chatButtons},
	}
	bot.telebot.Edit(msgToEdit, msg.String(), options)
}

func (bot *telegramBot) handleAddAdmin(ctx tb.Context, state State) {
	callback := ctx.Callback()
	_ = bot.telebot.Respond(callback)

	lang := ctx.Sender().LanguageCode
	msg := bot.bundle.T(lang, "Write the new admin ID.")

	_, _ = bot.telebot.Edit(callback.Message, msg)
	state.AddBotAdmin = true
	state.Save()
}
