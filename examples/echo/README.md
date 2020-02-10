# example: echo

`echo` is a basic Discord bot that has a simple `echo` command defined.

You can invoke this command with `e!echo <argument>`, or its alias `e!e <argument>`.

Additionally, there is an event handler bound which will be run on any message that is not a bot's message or itself.

## run

`go build`

`./echo -t <discord bot token>`