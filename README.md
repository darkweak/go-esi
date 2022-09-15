go-esi
------

go-esi is the implementation of the non-standard ESI (Edge-Side-Include) specification from the w3. With that you'll be able to use the ESI tags and process them in your favorite golang servers.

## What are the ESI tags
The ESI tags were introduced by Akamai to add some dynamic tags and only re-render these parts on the server-side.
The goal of that is to render only specific parts. For example, we want to render a full e-commerce webpage but only the cart is user-dependent. So we could render the "static" parts and store with a predefined TTL (e.g. 60 minutes), and only the cart would be requested to render the block.

There are multiple `esi` tags that we can use but the most used is the `esi:include` because that's the one to request another resource.

We can have many `esi:include` tags in a single response, and each `esi:include` tags can itself have one or more `esi:include` tags.

![esi page example](https://github.com/darkweak/go-esi/blob/master/docs/esi_2.jpg?sanitize=true)

We can have multiple `esi:include` tags in the page to request another resource and add its content to the main page.

![esi process example](https://github.com/darkweak/go-esi/blob/master/docs/esi_1.jpg?sanitize=true)

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
- [x] Træfik

### Caddy middleware
```bash
xcaddy build --with github.com/darkweak/go-esi/middleware/caddy
```
Refer to the [sample Caddyfile](https://github.com/darkweak/go-esi/blob/master/middleware/caddy/Caddyfile) to know how to configure that.

### Træfik middleware
```bash
# anywhere/traefik.yml
experimental:
  plugins:
    souin:
      moduleName: github.com/darkweak/go-esi
      version: v0.0.4
```
```bash
# anywhere/dynamic-configuration
http:
  routers:
    whoami:
      middlewares:
        - esi
      service: whoami
      rule: Host(`domain.com`)
  middlewares:
    esi:
      plugin:
        esi:
          # We don't care about the configuration but we have ot declare that block 
          # due to shitty træfik empty configuration handle.
          disable: false
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