# sayori

`sayori` is a dead simple command router based on [discordgo](https://github.com/bwmarrin/discordgo).

`sayori` uses no reflection and has no concept of subcommands, forming a 'flat' hierarchy. 

`Command` and `Event` are interfaces that describe entities that only run on a `MessageCreate` Discord event. 
A `Command` composes of an `Event` with prefix and alias matching, and argument parsing.
Parsing arguments in an `Event` is also optional.

You can bind a `Command` or `Event` to the router by plugging it into `.Command` or `.Event` and wrapping that with `.Has`, as shown in the example.
Filters can then be chained onto these handlers to control when the bound command handler fires. A given `Filter` will inspect the `MessageCreate` invocation context and if a match is found, will prevent the command handler from firing.

Alternatively, `.HandleDefault` is available if you want to implement your own handler that does not satisfy `Command` or `Event`.

More details on these interfaces are defined in `/router.go`.

To initialize `sayori`, a `Prefixer` must be defined. A `Prefixer` can load a prefix based on `guildID`
 or use a default prefix. This will only be used for parsing a `Command`.

```go
dg, err := discordgo.New("Bot " + Token)
if err != nil {
	fmt.Println("error creating Discord session,", err)
	return
}

router := sayori.New(dg, &Prefixer{})
router.Has(router.Command(&EchoCmd{}))

router.Has(router.Event(&OnMsg{}).
	Filter(sayori.MessagesBot).
	Filter(sayori.MessagesEmpty).
	Filter(sayori.MessagesSelf))

router.HasOnce(router.HandleDefault(func(_ *discordgo.Session, d *discordgo.MessageDelete) {
	log.Printf("A message was deleted: %v, %v, %v", d.Message.ID, d.Message.ChannelID, d.Message.GuildID)
}, nil))
```

## getting started

### installation

`go get github.com/pixeltopic/sayori`

See `/examples` for detailed usage.