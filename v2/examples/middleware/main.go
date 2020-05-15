package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pixeltopic/sayori/v2/context"

	"github.com/bwmarrin/discordgo"
	v2 "github.com/pixeltopic/sayori/v2"
)

// Variables used for command line parameters
var (
	Token string
)

const defaultPrefix = "?"

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

// Prefix loads a prefix
type Prefix struct{}

// Load returns the Default prefix no matter what the guildID given is.
func (p *Prefix) Load(_ string) (string, bool) { return p.Default(), true }

// Default returns the default router prefix
func (*Prefix) Default() string { return defaultPrefix }

// Validator validates if a user has admin privilege
type Validator struct{}

// Do a check for valid permissions
func (*Validator) Do(ctx *context.Context) error {
	aPerm, err := ctx.Ses.State.UserChannelPermissions(
		ctx.Msg.Author.ID, ctx.Msg.ChannelID)
	if err != nil {
		return err
	}

	if aPerm&discordgo.PermissionAdministrator != discordgo.PermissionAdministrator {
		return errors.New("you don't have admin perms :(")
	}

	return nil
}

// Privilege is a privileged command only admins can use.
type Privilege struct{}

// Handle handles the command
func (*Privilege) Handle(ctx *context.Context) error {
	_, _ = ctx.Ses.ChannelMessageSend(ctx.Msg.ChannelID, "You are privileged!")
	return nil
}

// Resolve handles any errors
func (*Privilege) Resolve(ctx *context.Context) {
	if ctx.Err != nil {
		_, _ = ctx.Ses.ChannelMessageSend(ctx.Msg.ChannelID, ctx.Err.Error())
	}
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	router := v2.New(dg)

	router.Has(
		v2.NewRoute(&Prefix{}).
			On("p", "priv", "privileged").
			Do(&Privilege{}).
			Use(&Validator{}),
	)

	err = router.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	_ = router.Close()
}
