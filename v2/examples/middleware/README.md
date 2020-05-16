# example: middleware

`middleware` is a basic Discord bot that has a several middlewares defined, along with a simple command containing the aliases `?p`, `?priv`, and `?privileged`.
The command is explicitly disallowed from being invoked in a private DM channel thanks to the `Filter` middleware and Sayori's `Filter` package.

The `Validate` middleware will be run after `Filter` and will determine whether the user calling it has admin privileges.

These middlewares are executed in the order they are added to the route. In the following two examples, the `Filter` middleware will always be run before `Validate`.
```go
router.Has(
    v2.NewRoute(&Prefix{}).
        On("p", "priv", "privileged").
        Do(&Privilege{}).
        Use(&Filter{}, &Validate{}),
)
```
```go
router.Has(
    v2.NewRoute(&Prefix{}).
        On("p", "priv", "privileged").
        Do(&Privilege{}).
        Use(&Filter{}).Use(&Validate{}),
)
```

## run

`go build`

`./middleware -t <discord bot token>`