package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pixeltopic/sayori/v2/context"

	"github.com/bwmarrin/discordgo"
	v2 "github.com/pixeltopic/sayori/v2"
)

// Variables used for command line parameters
var (
	Token string
)

const (
	defaultPrefix = "^"
	alias         = "foo bar"
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

type tokLengthErr struct{}

func (*tokLengthErr) Error() string { return "not enough tokens to parse message" }

// Prefix loads a prefix
type Prefix struct{}

// Load returns the Default prefix no matter what the guildID given is.
func (p *Prefix) Load(_ string) (string, bool) { return p.Default(), true }

// Default returns the default router prefix
func (*Prefix) Default() string { return defaultPrefix }

// Parse is a command that will attempt to parse an alias that has a whitespace
type Parse struct{}

// Parse joins the first 2 tokens, returns error if under 2 tokens
func (*Parse) Parse(s string) ([]string, error) {
	toks := strings.Fields(s)
	if len(toks) < 2 {
		return []string{}, &tokLengthErr{}
	}

	first, second := toks[0], toks[1]

	// start of message + length of token + space - start of second token == 0
	dist := (strings.Index(s, first) + len(first) + 1) - strings.Index(s, second)

	if dist != 0 {
		return toks, nil
	}

	newToks := []string{strings.Join(toks[:2], " ")}
	newToks = append(newToks, toks[2:]...)

	return newToks, nil
}

// Handle handles the command
func (*Parse) Handle(ctx *context.Context) error {
	_, _ = ctx.Ses.ChannelMessageSend(ctx.Msg.ChannelID, "Custom parser ran!")
	return nil
}

// Resolve handles any errors
func (*Parse) Resolve(ctx *context.Context) {
	switch err := ctx.Err.(type) {
	case *tokLengthErr:
		_, _ = ctx.Ses.ChannelMessageSend(ctx.Msg.ChannelID, err.Error())
	default:
	}
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	router := v2.New(dg)

	router.Has(v2.NewRoute(&Prefix{}).On(alias).Do(&Parse{}))

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
