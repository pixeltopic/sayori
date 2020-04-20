# example: middleware

`middleware` is a basic Discord bot that has a simple `Checker` middleware defined.

This middleware will be run when the command `.p` is invoked, and will determine whether the user calling it has admin privileges.

## run

`go build`

`./middlewares -t <discord bot token>`