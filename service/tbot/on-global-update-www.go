package tbot

import (
	"html"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	tb "gopkg.in/tucnak/telebot.v3"
)

// onGlobalUpdateWWW updates the links on web page on /updatewww command.
func (bot *telegramBot) onGlobalUpdateWWW(ctx tb.Context, settings chatSettings) {
	lang := ctx.Sender().LanguageCode

	if bot.gitTemporaryDir == "" || bot.gitSSHKey == "" {
		_ = ctx.Send(bot.bundle.T(lang, "Website updater not configured"))
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Failed to update website on /updatewww: configuration is missing")
		return
	}
	gitTempDir := filepath.Join(bot.gitTemporaryDir, "gittmp")

	chat := ctx.Chat()
	if chat == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Chat, ignored")
		return
	}

	// First, prepare the web page content
	msg, err := bot.telebot.Send(chat, "⚙️  " + bot.bundle.T(lang, "Prepare group list"))
	if err != nil {
		bot.logger.WithError(err).Error("Failed to send message")
		return
	}
	linksPageContent, err := bot.prepareLinksWebPageContent()
	if err != nil {
		_, _ = bot.telebot.Edit(msg, "❌ " + bot.bundle.T(lang, "Prepare group list") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to prepare the group list for website update")
		return
	}

	// Create a temporary directory (if it doesn't exist), or remove its content
	if err := os.Mkdir(gitTempDir, 0750); err != nil {
		_, _ = bot.telebot.Edit(msg, "❌ " + bot.bundle.T(lang, "Prepare group list") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to create temporary directory for website update")
		return
	}
	if err := removeContents(gitTempDir); err != nil {
		_, _ = bot.telebot.Edit(msg, "❌ " + bot.bundle.T(lang, "Prepare group list") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to clean up the temporary directory")
		return
	}

	// Prepare SSH Authentication
	pubkeys, err := ssh.NewPublicKeysFromFile("git", bot.gitSSHKey, bot.gitSSHKeyPassphrase)
	if err != nil {
		_, _ = bot.telebot.Edit(msg, "❌ " + bot.bundle.T(lang, "Prepare group list") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to load SSH keys")
		return
	}

	// Clone the repo...
	msg, err = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n⚙️  " + bot.bundle.T(lang, "Cloning"))
	if err != nil {
		bot.logger.WithError(err).Error("Failed to edit message")
		return
	}
	r, err := git.PlainClone(gitTempDir, false, &git.CloneOptions{
		// TODO: Make this configurable.
		URL:  "git@gitlab.com:sapienzastudents/sapienzahub.git",
		Auth: pubkeys,
	})
	if err != nil {
		_, _ = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n❌ " + bot.bundle.T(lang, "Cloning") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to clone SSH repo")
		return
	}

	// ...from origin and...
	remote, err := r.Remote("origin")
	if err != nil {
		_, _ = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n❌ " + bot.bundle.T(lang, "Cloning") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to get origin remote")
		return
	}

	// ...all refs, including HEAD.
	opts := &git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
	}
	if err := remote.Fetch(opts); err != nil {
		_, _ = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n❌ " + bot.bundle.T(lang, "Cloning") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to fetch updates from remote")
		return
	}

	w, err := r.Worktree()
	if err != nil {
		_, _ = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n❌ " + bot.bundle.T(lang, "Cloning") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to get the working tree")
		return
	}

	// Checkout the master branch
	branchRefName := plumbing.NewBranchReferenceName("master")
	if err := w.Checkout(&git.CheckoutOptions{Branch: branchRefName}); err != nil {
		_, _ = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n❌ " + bot.bundle.T(lang, "Cloning") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to checkout the branch")
		return
	}

	// Overwrite the file
	msg, _ = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n✅ " + bot.bundle.T(lang, "Cloning") + "\n⚙️  " + bot.bundle.T(lang, "File creation"))
	err = ioutil.WriteFile(filepath.Join(gitTempDir, "content", "social.md"), []byte(linksPageContent), 0600)
	if err != nil {
		_, _ = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n✅ " + bot.bundle.T(lang, "Cloning") + "\n❌ " + bot.bundle.T(lang, "File creation") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to write to file")
		return
	}

	// Add the file to the next git commit
	if _, err := w.Add(filepath.Join("content", "social.md")); err != nil {
		_, _ = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n✅ " + bot.bundle.T(lang, "Cloning") + "\n❌ " + bot.bundle.T(lang, "File creation") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to add the modified file to the git repo")
		return
	}

	msg, err = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n✅ " + bot.bundle.T(lang, "Cloning") + "\n✅ " + bot.bundle.T(lang, "File creation") + "\n⚙️  " + bot.bundle.T(lang, "Commit and push"))
	if err != nil {
		bot.logger.WithError(err).Error("Failed to edit message")
		return
	}

	// Commit changes.
	_, err = w.Commit("Update social groups links", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "SapienzaStudentsBot",
			Email: "sapienzastudentsbot@domain.invalid",
			When:  time.Now(),
		},
		All: true,
	})
	if err != nil {
		_, _ = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n✅ " + bot.bundle.T(lang, "Cloning") + "\n✅ " + bot.bundle.T(lang, "File creation") + "\n❌️ " + bot.bundle.T(lang, "Commit and push") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to commit to repo")
		return
	}

	// Push the commit to the origin repository
	if err = r.Push(&git.PushOptions{RemoteName: "origin"}); err != nil {
		_, _ = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n✅ " + bot.bundle.T(lang, "Cloning") + "\n✅ " + bot.bundle.T(lang, "File creation") + "\n❌️ " + bot.bundle.T(lang, "Commit and push") + "\n\n" + err.Error())
		bot.logger.WithError(err).Error("Failed to push to remote origin")
		return
	}
	_, err = bot.telebot.Edit(msg, "✅ " + bot.bundle.T(lang, "Prepare group list") + "\n✅ " + bot.bundle.T(lang, "Cloning") + "\n✅ " + bot.bundle.T(lang, "File creation") + "\n⚙️  " + bot.bundle.T(lang, "Commit and push") + "\n✅️ " + bot.bundle.T(lang, "Pushed"))
	if err != nil {
		bot.logger.WithError(err).Error("Failed to edit message")
		return
	}

	// GitLab automatically publish the update website.
}

// prepareLinksWebPageContent returns the new markdown content for the group
// links web page.
//
// TODO: Rewrite all in HTML?
func (bot *telegramBot) prepareLinksWebPageContent() (string, error) {
	// Get all categories
	categories, err := bot.db.GetChatTree()
	if err != nil {
		return "", err
	}

	// Page header
	msg := strings.Builder{}
	msg.WriteString(`+++
description = "Pagina contenenti link ai gruppi social"
title = "Link gruppi social"
type = "post"
date = "`)
	msg.WriteString(time.Now().Format("2006-01-02"))
	msg.WriteString(`"
+++

# Gruppi Telegram

Qui di seguito trovi un indice di gruppi di studenti della Sapienza su Telegram. **Clicca su START dopo che si è aperto Telegram per avere il link**.

Se vuoi aggiungere il tuo gruppo, segui le [indicazioni in questa pagina!](/social_add/)

`)

	// Write categories and subcategories
	for _, category := range categories.GetSubCategoryList() {
		var l1cat = categories.SubCategories[category]
		if len(l1cat.Chats) == 0 && len(l1cat.GetSubCategoryList()) == 0 {
			continue
		}

		msg.WriteString("\n## ")
		msg.WriteString(html.EscapeString(category))
		msg.WriteString("\n")

		for _, v := range l1cat.GetChats() {
			_ = bot.printChatsInMarkdown(&msg, v)
		}
		for _, subcat := range l1cat.GetSubCategoryList() {
			msg.WriteString("\n### ")
			msg.WriteString(html.EscapeString(subcat))
			msg.WriteString("\n")
			var l2cat = l1cat.SubCategories[subcat]
			for _, v := range l2cat.GetChats() {
				_ = bot.printChatsInMarkdown(&msg, v)
			}
		}
	}

	if len(categories.Chats) > 0 {
		msg.WriteString("## Senza categoria\n")

		for _, v := range categories.GetChats() {
			_ = bot.printChatsInMarkdown(&msg, v)
		}
	}

	return msg.String(), nil
}

// printChatsInMarkdown appends a line for the chat in Markdown format to the
// given message buffer.
func (bot *telegramBot) printChatsInMarkdown(msg *strings.Builder, v *tb.Chat) error {
	settings, err := bot.getChatSettings(v)
	if err != nil {
		bot.logger.WithError(err).Error("Failed to get chatroom config")
		return err
	}
	if settings.Hidden {
		return nil
	}

	chatUUID, err := bot.db.GetUUIDFromChat(v.ID)
	if err != nil {
		return err
	}

	msg.WriteString("* [")
	msg.WriteString(v.Title)
	msg.WriteString("](")
	msg.WriteString("https://telegram.me/SapienzaStudentsBot?start=" + chatUUID.String())
	msg.WriteString(")\n")
	return nil
}
