package main

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/pixeltopic/sayori"
)

// AdminMiddleware checks if a user has admin privilege
type AdminMiddleware struct{}

// Do a check for valid permissions
func (*AdminMiddleware) Do(ctx sayori.Context) error {
	aperm, err := ctx.Session.State.UserChannelPermissions(ctx.Message.Author.ID, ctx.Message.ChannelID)
	if err != nil {
		return err
	}

	if aperm&discordgo.PermissionAdministrator != discordgo.PermissionAdministrator {
		return errors.New("you don't have admin perms :(")
	}

	return nil
}
