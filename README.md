# sayori

`sayori` is a dead simple command router based on [discordgo](https://github.com/bwmarrin/discordgo).

`sayori` uses no reflection and has no concept of subcommands, forming a 'flat' hierarchy. 

`Command` and `Event` define interfaces that only run on a `MessageCreate`. 
An `Event` is essentially a `Command`, but skips prefix and alias matching.
Parsing arguments in an `Event` is also optional.

If a handler that does not satisfy `Event` and `Command` is added via `Has`, it will auto-default to `discordgo`'s handler types.

More details on these interfaces are defined in `/router.go`.

To initialize `sayori`, a `Prefixer` must be defined. A `Prefixer` can load a prefix based on `guildID`
 or use a default prefix. This will only be used for parsing a `Command`.

```go
router := sayori.New(dgoSession, &Prefixer{})
router.Has(&EchoCmd{}, nil)
router.Has(func(_ *discordgo.Session, d *discordgo.MessageDelete) {
	log.Printf("A message was deleted: %v, %v, %v", d.Message.ID, d.Message.ChannelID, d.Message.GuildID)
}, nil)
```

## getting started

### installation

`go get github.com/pixeltopic/sayori`

See `/examples` for detailed usage.