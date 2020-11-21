package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"html"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func onGlobalUpdateWww(m *tb.Message, _ botdatabase.ChatSettings) {
	gitTempDir := os.Getenv("GIT_TEMP_DIR")
	gitSshKeyFile := os.Getenv("GIT_SSH_KEY")
	if gitTempDir == "" || gitSshKeyFile == "" {
		_, _ = b.Send(m.Chat, "Website updater not configured")
		logger.Warning("Website update requested but the configuration is missing")
		return
	}
	gitTempDir = filepath.Join(gitTempDir, "gittmp")

	// Prepare the group list
	msg, _ := b.Send(m.Chat, "⚙️ Prepare group list")
	groupList, err := prepareGroupListForWeb()
	if err != nil {
		_, _ = b.Edit(msg, "❌ Prepare group list\n\n"+err.Error())
		logger.WithError(err).Error("can't prepare the group list for website update")
		return
	}

	// Creating temp dir
	_ = os.Mkdir(gitTempDir, 0750)
	err = RemoveContents(gitTempDir)
	if err != nil {
		_, _ = b.Edit(msg, "❌ Prepare group list\n\n"+err.Error())
		logger.WithError(err).Error("can't clean up the temp directory")
		return
	}

	// Authentication
	pubkeys, err := ssh.NewPublicKeysFromFile("git", gitSshKeyFile, os.Getenv("GIT_SSH_KEY_PASS"))
	if err != nil {
		_, _ = b.Edit(msg, "❌ Prepare group list\n\n"+err.Error())
		logger.WithError(err).Error("can't load SSH keys")
		return
	}

	// Clone the repo and checkout the branch
	msg, _ = b.Edit(msg, "✅ Prepare group list\n⚙️ Cloning")
	r, err := git.PlainClone(gitTempDir, false, &git.CloneOptions{
		URL:  "git@gitlab.com:sapienzastudents/sapienzahub.git",
		Auth: pubkeys,
	})
	if err != nil {
		_, _ = b.Edit(msg, "✅ Prepare group list\n❌ Cloning\n\n"+err.Error())
		logger.WithError(err).Error("can't clone SSH repo")
		return
	}

	remote, err := r.Remote("origin")
	if err != nil {
		_, _ = b.Edit(msg, "✅ Prepare group list\n❌ Cloning\n\n"+err.Error())
		logger.WithError(err).Error("can't get origin remote")
		return
	}

	opts := &git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
	}

	if err := remote.Fetch(opts); err != nil {
		_, _ = b.Edit(msg, "✅ Prepare group list\n❌ Cloning\n\n"+err.Error())
		logger.WithError(err).Error("can't fetch updates from remote")
		return
	}

	w, err := r.Worktree()
	if err != nil {
		_, _ = b.Edit(msg, "✅ Prepare group list\n❌ Cloning\n\n"+err.Error())
		logger.WithError(err).Error("can't get the working tree")
		return
	}
	branchRefName := plumbing.NewBranchReferenceName("master")
	err = w.Checkout(&git.CheckoutOptions{
		Branch: branchRefName,
	})
	if err != nil {
		_, _ = b.Edit(msg, "✅ Prepare group list\n❌ Cloning\n\n"+err.Error())
		logger.WithError(err).Error("can't get checkout the branch")
		return
	}

	// Update the file
	msg, _ = b.Edit(msg, "✅ Prepare group list\n✅ Cloning\n⚙️ Create file")
	err = ioutil.WriteFile(filepath.Join(gitTempDir, "content", "social.md"), []byte(groupList), 0600)
	if err != nil {
		_, _ = b.Edit(msg, "✅ Prepare group list\n✅ Cloning\n❌ Create file\n\n"+err.Error())
		logger.WithError(err).Error("can't write to file")
		return
	}

	_, err = w.Add(filepath.Join("content", "social.md"))
	if err != nil {
		_, _ = b.Edit(msg, "✅ Prepare group list\n✅ Cloning\n❌ Create file\n\n"+err.Error())
		logger.WithError(err).Error("can't add the modified file to the git repo")
		return
	}

	// Commit
	msg, _ = b.Edit(msg, "✅ Prepare group list\n✅ Cloning\n✅ Create file\n⚙️ Commit&push")
	_, err = w.Commit("Update social groups links", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "SapienzaStudentsBot",
			Email: "sapienzastudentsbot@domain.invalid",
			When:  time.Now(),
		},
		All: true,
	})
	if err != nil {
		_, _ = b.Edit(msg, "✅ Prepare group list\n✅ Cloning\n✅ Create file\n❌️ Commit&push\n\n"+err.Error())
		logger.WithError(err).Error("can't commit to repo")
		return
	}

	// Push
	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
	})
	if err != nil {
		_, _ = b.Edit(msg, "✅ Prepare group list\n✅ Cloning\n✅ Create file\n❌️ Commit&push\n\n"+err.Error())
		logger.WithError(err).Error("can't push to remote origin")
		return
	}
	_, _ = b.Edit(msg, "✅ Prepare group list\n✅ Cloning\n✅ Create file\n✅️ Commit&push\n✅️ Pushed")
}

func prepareGroupListForWeb() (string, error) {
	// Get all categories
	categories, err := botdb.GetChatTree(b)
	if err != nil {
		return "", err
	}

	msg := strings.Builder{}
	msg.WriteString("+++\ndescription = \"Pagina contenenti link ai gruppi social\"\ntitle = \"Link gruppi social\"\ntype = \"post\"\ndate = \"")
	msg.WriteString(time.Now().Format("2006-01-02"))
	msg.WriteString("\"\n+++\n\n# Gruppi Telegram\n\n")

	msg.WriteString("## Gruppi generali Sapienza\n")
	for _, v := range categories.Chats {
		_ = printChatsInMarkdown(&msg, v)
	}

	for _, category := range categories.GetSubCategoryList() {
		msg.WriteString("\n## ")
		msg.WriteString(html.EscapeString(category))
		msg.WriteString("\n")

		var l1cat = categories.SubCategories[category]
		for _, v := range l1cat.Chats {
			_ = printChatsInMarkdown(&msg, v)
		}
		for _, subcat := range l1cat.GetSubCategoryList() {
			msg.WriteString("\n### ")
			msg.WriteString(html.EscapeString(subcat))
			msg.WriteString("\n")
			var l2cat = l1cat.SubCategories[subcat]
			for _, v := range l2cat.Chats {
				_ = printChatsInMarkdown(&msg, v)
			}
		}
	}

	return msg.String(), nil
}

func printChatsInMarkdown(msg *strings.Builder, v *tb.Chat) error {
	settings, err := botdb.GetChatSetting(b, v)
	if err != nil {
		logger.WithError(err).Error("Error getting chatroom config")
		return err
	}
	if settings.Hidden {
		return nil
	}

	if v.InviteLink == "" {
		v.InviteLink, err = b.GetInviteLink(v)

		if err != nil && err.Error() == tb.ErrGroupMigrated.Error() {
			apierr, _ := err.(*tb.APIError)
			v, err = b.ChatByID(fmt.Sprint(apierr.Parameters["migrate_to_chat_id"]))
			if err != nil {
				logger.Warning("can't get chat info ", err)
				return err
			}

			v.InviteLink, err = b.GetInviteLink(v)
			if err != nil {
				logger.Warning("can't get invite link ", err)
				return err
			}
		} else if err != nil {
			logger.Warning("can't get chat info ", err)
			return err
		}
		_ = botdb.UpdateMyChatroomList(v)
	}

	msg.WriteString("* [")
	msg.WriteString(v.Title)
	msg.WriteString("](")
	msg.WriteString(v.InviteLink)
	msg.WriteString(")\n")
	return nil
}
