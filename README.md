# sayori

`sayori` is a dead simple command router based on [discordgo](https://github.com/bwmarrin/discordgo).

`sayori` uses no reflection and has no concept of subcommands, forming a 'flat' hierarchy. 

`Command` and `Event` only run on a `MessageCreate`. An `Event` is essentially a `Command` but with less restrictions.
More details on these interfaces are defined in `/router.go`.

To initialize `sayori`, a `Prefixer` must be defined. A `Prefixer` can load a prefix based on `guildID`
 or use a default prefix. This will only be used for parsing a `Command`.

```go
router := sayori.New(dgoSession, &Prefixer{})
router.Has(&EchoCmd{}, nil) // Command type
router.Will(&OnMsg{}, nil) // Event type
```

## getting started

### installation

`go get github.com/pixeltopic/sayori`

See `/examples` for detailed usage.