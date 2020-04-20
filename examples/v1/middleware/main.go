package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/pixeltopic/sayori"
)

// Variables used for command line parameters
var (
	Token string
)

const defaultPrefix = "."

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

// Checker checks if a user has admin privilege
type Checker struct{}

// Do a check for valid permissions
func (*Checker) Do(ctx sayori.Context) error {
	aperm, err := ctx.Session.State.UserChannelPermissions(
		ctx.Message.Author.ID, ctx.Message.ChannelID)
	if err != nil {
		return err
	}

	if aperm&discordgo.PermissionAdministrator != discordgo.PermissionAdministrator {
		return errors.New("you don't have admin perms :(")
	}

	return nil
}

// Privilege is a privileged command only admins can use.
type Privilege struct{}

// Match examines the first token to see if it matches a valid alias
func (*Privilege) Match(toks sayori.Toks) (string, bool) {
	alias, ok := toks.Get(0)
	if !ok {
		return "", false
	}
	alias = strings.ToLower(alias)

	for _, validAlias := range []string{"p", "priv", "privileged"} {
		if alias == validAlias {
			return alias, true
		}
	}
	return "", false
}

// Handle handles the echo command
func (*Privilege) Handle(ctx sayori.Context) error {
	_, _ = ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "You are privileged!")
	return nil
}

// Resolve handles any errors
func (*Privilege) Resolve(ctx sayori.Context) {
	if ctx.Err != nil {
		_, _ = ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Err.Error())
	}
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	router := sayori.New(dg, &Prefix{})

	router.Has(router.Command(&Privilege{}).Use(&Checker{}))

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
