go-esi
------

go-esi is the implementation of the non-standard ESI (Edge-Side-Include) specification from the w3. With that you'll be able to use the ESI tags and process them in your favorite golang servers.

## References
https://www.w3.org/TR/esi-lang/

## Install
```bash
go get -u github.com/darkweak/go-esi
```

## Usage
```go
import (
    // ...
    github.com/darkweak/go-esi/esi
)

//...

func functionToParseESITags(b []byte, r *http.Request) []byte {
    // Returns the parsed response.
    res := esi.Parse(b, r)

    //...
    return res
}
```

## Available as middleware
- [x] Caddy

### Caddy middleware
```bash
xcaddy build --with github.com/darkweak/go-esi/middleware/caddy
```
Refer to the [sample Caddyfile](https://github.com/darkweak/go-esi/blob/master/middleware/caddy/Caddyfile) to know how to configure that.

## TODO
- [x] choose tag
- [x] comment tag
- [x] escape tag
- [x] include tag
- [x] remove tag
- [x] otherwise tag
- [ ] try tag
- [x] vars tag
- [x] when tag