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

// Prefixer loads a prefix
type Prefixer struct {
}

// Load returns a prefix based on the serverID
func (p *Prefixer) Load(serverID string) (string, bool) {
	return p.Default(), true
}

// Default returns the default router prefix
func (*Prefixer) Default() string {
	return "e!"
}

// EchoCmd defines a simple EchoCmd.
type EchoCmd struct {
	aliases []string
}

// Match returns a matched alias if the bool is true
func (c *EchoCmd) Match(fullcommand string) (string, bool) {
	for i := range c.aliases {
		if strings.HasPrefix(fullcommand, c.aliases[i]) {
			return c.aliases[i], true
		}
	}
	return "", false
}

// Parse returns args found in the command
func (c *EchoCmd) Parse(fullcommand string) sayori.Args {
	sslice := strings.Fields(fullcommand)
	if len(sslice) < 2 {
		return sayori.Args{
			"error": errors.New("not enough args to echo :("),
		}
	}
	return sayori.Args{
		"alias":   sslice[0],
		"to-echo": strings.Join(sslice[1:], " "),
	}
}

// Handle handles the echo command
func (c *EchoCmd) Handle(ctx sayori.Context) {
	if ctx.Message.Author.ID == ctx.Session.State.User.ID {
		return
	}
	if err, ok := ctx.Args["error"]; ok {
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, err.(error).Error())
	}
	if msg, ok := ctx.Args["to-echo"]; ok {
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "Echoing! "+msg.(string))
	}
}

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	router := sayori.NewSessionRouter(dg, &Prefixer{})
	router.MessageHandler(&EchoCmd{
		aliases: []string{"echo", "e"},
	}, sayori.NewEventHandlerRule())

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}
