# sayori

Sayori is a dead simple command router based on [discordgo](https://github.com/bwmarrin/discordgo).
Typically, command syntax for Discord Bots look something like `[prefix][alias] [args]`.
Sayori makes this easy by breaking down a command into different components that will be used for handling.

V1 of Sayori does not include official subcommand support.

## Getting Started

### Installing

You can install the latest release of Sayori by using:

```
go get github.com/pixeltopic/sayori
```

Then include Sayori in your application:

```go
import "github.com/pixeltopic/sayori"
```

### Usage

`Command` and `Event` are interfaces that describe entities that only run on a `discordgo.MessageCreate` event. 
- A `Command` composes of an `Event` with prefix and alias matching, and argument parsing.
- Parsing arguments in an `Event` is also optional.
- A `Filter` can be chained onto the `Event` or `Command` when binding it to the router.
  - They're essentially a middleware that can filter out messages from executing the handler if they meet a certain criteria.

Sayori also allows you to implement bind custom handlers that do not satisfy `Command` or `Event`.
These behave like discordgo's handlers.

More details on these interfaces are defined in `/router.go`.

To initialize Sayori, a `Prefixer` must be defined.
- A `Prefixer` can load a command prefix based on `guildID` or use a default prefix. 
- This will only be used for parsing a `Command`, as events naturally would not consider a prefix.

```go
dg, err := discordgo.New("Bot " + Token)
if err != nil {
	fmt.Println("error creating Discord session,", err)
	return
}

router := sayori.New(dg, &Prefixer{})
router.Has(router.Command(&EchoCmd{}))

// bind an Event handler that runs on every message except those 
// that are sent by bots, have no body, or are sent by the bot session itself
router.Has(router.Event(&OnMsg{}).
	Filter(sayori.MessagesBot).
	Filter(sayori.MessagesEmpty).
	Filter(sayori.MessagesSelf))

// bind a discordgo handler function to the router
router.HasOnce(router.HandleDefault(func(_ *discordgo.Session, d *discordgo.MessageDelete) {
	log.Printf("A message was deleted: %v, %v, %v", d.Message.ID, d.Message.ChannelID, d.Message.GuildID)
}, nil))
```

See `/examples` for detailed usage.


## License

This project is licensed under the BSD 3-Clause License - see the [LICENSE.md](https://github.com/pixeltopic/sayori/blob/master/LICENSE) file for details

## Acknowledgments

* Inspired by rfrouter and dgrouter.
