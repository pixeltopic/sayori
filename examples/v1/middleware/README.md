# example: middlewares

`middlewares` is a basic Discord bot that has a simple `AdminMiddleware` middleware defined.

This middleware will be run when the command `.which` is invoked, and will determine whether the user calling it has admin privileges.

## run

`go build`

`./middlewares -t <discord bot token>`