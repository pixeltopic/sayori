# example: echo

`echo` is a Discord bot that has a several commands and subcommands defined.

## commands

`e!echo` which echoes text following this command. It has the subcommands `fmt` and `color`.
The subcommand `fmt` also has the subcommand `color`.

This means a commands like so are valid:

```
e!echo fmt color blah blah blah
e!echo fmt blah blah blah
e!echo color blah blah blah
e!echo blah blah blah
```

`e!fmt` and `e!color` are also valid commands.
`fmt` has the subcommand `color`, but `color` has no subcommands.

So these commands are valid:

```
e!fmt color blah blah blah
e!fmt blah blah blah
e!color blah blah blah
```

## run

`go build`

`./echo -t <discord bot token>`