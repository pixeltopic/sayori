package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pixeltopic/sayori/v2/filter"

	"context"

	"github.com/bwmarrin/discordgo"
	sayori "github.com/pixeltopic/sayori/v2"
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

// Filter will filter out invocations from a private DM channel
type Filter struct{}

// Do a check for valid invocation context
func (*Filter) Do(ctx context.Context) error {
	f := filter.New(filter.MsgIsPrivate)

	valid, _ := f.Validate(ctx)
	if !valid {
		return errors.New("invalid channel :(")
	}

	return nil
}

// Validate validates if a user has admin privilege
type Validate struct{}

// Do a check for valid permissions
func (*Validate) Do(ctx context.Context) error {

	cmd := sayori.CmdFromContext(ctx)

	aPerm, err := cmd.Ses.State.UserChannelPermissions(
		cmd.Msg.Author.ID, cmd.Msg.ChannelID)
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
func (*Privilege) Handle(ctx context.Context) error {
	cmd := sayori.CmdFromContext(ctx)

	_, _ = cmd.Ses.ChannelMessageSend(cmd.Msg.ChannelID, "You are privileged!")
	return nil
}

// Resolve handles any errors
func (*Privilege) Resolve(ctx context.Context) {
	cmd := sayori.CmdFromContext(ctx)

	if cmd.Err != nil {
		_, _ = cmd.Ses.ChannelMessageSend(cmd.Msg.ChannelID, cmd.Err.Error())
	}
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	router := sayori.New(dg)

	router.Has(
		sayori.NewRoute(&Prefix{}).
			On("p", "priv", "privileged").
			Do(&Privilege{}).
			Use(&Filter{}, &Validate{}),
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
