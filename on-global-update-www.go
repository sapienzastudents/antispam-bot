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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
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
	chatrooms, err := botdb.ListMyChatrooms()
	if err != nil {
		logger.WithError(err).Error("Error getting chatroom list")
		return "", err
	} else {
		sort.Slice(chatrooms, func(i, j int) bool {
			return chatrooms[i].Title < chatrooms[j].Title
		})

		msg := strings.Builder{}
		msg.WriteString("+++\ndescription = \"Pagina contenenti link ai gruppi social\"\ntitle = \"Link gruppi social\"\ntype = \"post\"\ndate = \"")
		msg.WriteString(time.Now().Format("2006-01-02"))
		msg.WriteString("\"\n+++\n\n# Gruppi Telegram\n\n")

		for _, v := range chatrooms {
			settings, err := botdb.GetChatSetting(b, v)
			if err != nil {
				logger.WithError(err).Error("Error getting chatroom config")
				continue
			}
			if settings.Hidden {
				continue
			}

			if v.InviteLink == "" {
				v.InviteLink, err = b.GetInviteLink(v)

				if err != nil && err.Error() == tb.ErrGroupMigrated.Error() {
					apierr, _ := err.(*tb.APIError)
					v, err = b.ChatByID(fmt.Sprint(apierr.Parameters["migrate_to_chat_id"]))
					if err != nil {
						logger.Warning("can't get chat info ", err)
						continue
					}

					v.InviteLink, err = b.GetInviteLink(v)
					if err != nil {
						logger.Warning("can't get invite link ", err)
						continue
					}
				} else if err != nil {
					logger.Warning("can't get chat info ", err)
					continue
				}
				_ = botdb.UpdateMyChatroomList(v)
			}

			msg.WriteString("* [")
			msg.WriteString(v.Title)
			msg.WriteString("](")
			msg.WriteString(v.InviteLink)
			msg.WriteString(")\n")
		}

		return msg.String(), nil
	}
}
