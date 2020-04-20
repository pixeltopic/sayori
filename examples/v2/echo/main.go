package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	v2 "github.com/pixeltopic/sayori/v2"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func onDelete(_ *discordgo.Session, d *discordgo.MessageDelete) {
	log.Printf("A message was deleted: %v, %v, %v", d.Message.ID, d.Message.ChannelID, d.Message.GuildID)
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	p := &Prefixer{}

	router := v2.New(dg)

	router.Has(
		v2.NewRoute(p).Do(&EchoCmd{}).On("echo", "e").Has(
			v2.NewRoute(nil).Do(&EchoSubCmd{}).On("f", "fmt")))

	router.Has(v2.NewRoute(nil).Do(&OnMsg{}))

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
