package tbot

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"reflect"
)

var botPermissionsTag = map[string]string{
	"can_change_info":      "C",
	"can_delete_messages":  "D",
	"can_invite_users":     "I",
	"can_restrict_members": "R",
	"can_pin_messages":     "N",
	"can_promote_members":  "P",
}

var botPermissionsText = map[string]string{
	"can_change_info":      "Change group info",
	"can_delete_messages":  "Delete messages",
	"can_invite_users":     "Invite users via link",
	"can_restrict_members": "Restrict/ban users",
	"can_pin_messages":     "Pin messages",
	"can_promote_members":  "Add new admins",
}

// Do not use this array to check if a permission is granted or not, use ChatMember fields
func synthetizePrivileges(user *tb.ChatMember) []string {
	var ret []string
	t := reflect.TypeOf(user.Rights)
	right := reflect.ValueOf(user.Rights)
	for i := 0; i < t.NumField(); i++ {
		k := t.Field(i).Tag.Get("json")
		_, ok := botPermissionsTag[k]
		if !ok {
			// Skip this field
			continue
		}

		f := right.Field(i)
		if !f.Bool() {
			ret = append(ret, k)
		}
	}
	return ret
}
