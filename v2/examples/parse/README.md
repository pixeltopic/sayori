# example: parse

`parse` is a basic Discord bot demonstrates Sayori's custom message parsing support.

When `^foo bar` is invoked, it will match with an alias. However, `^foo  bar`, `^foo` for example will not get matched.

Any errors that occur in parsing can be handled by `Resolve`. Note that if you have multiple commands which produce parse errors with a single invocation, _all_ of their respective `Resolve` handlers will be run.

## run

`go build`

`./parse -t <discord bot token>`