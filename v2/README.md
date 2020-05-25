# Sayori

Sayori is a dead simple command router based on [discordgo](https://github.com/bwmarrin/discordgo) with

- custom guild prefix support
- message parsing
- multiple command aliases
- subcommands
- middlewares

## Getting Started

### Installing

You can install the latest release of Sayori by using:

```
go get github.com/pixeltopic/sayori/v2
```

Then include Sayori in your application:

```
import sayori "github.com/pixeltopic/sayori/v2"
```

### Usage

```go
package main

import (
	"github.com/bwmarrin/discordgo"
	sayori "github.com/pixeltopic/sayori/v2"
)

func main() {
    dg, err := discordgo.New("Bot " + "<mydiscordtoken>")
    if err != nil {
        return
    }
    
    router := sayori.New(dg)
    
    echo := sayori.NewRoute(&Prefix{}).Do(&Echo{}).On("echo", "e")
    
    router.Has(echo)
    
    err = router.Open()
    if err != nil {
        return
    }
}
```

See `/examples` for detailed usage.


## License

This project is licensed under the BSD 3-Clause License - see the [LICENSE.md](https://github.com/pixeltopic/sayori/blob/master/LICENSE) file for details

## Acknowledgments

* Inspired by rfrouter and dgrouter.
