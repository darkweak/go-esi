go-esi
------

go-esi is the implementation of the non-standard ESI (Edge-Side-Include) specification from the w3. With that you'll be able to use the ESI tags and process them in your favorite golang servers.

## References
https://www.w3.org/TR/esi-lang/

## Install
```bash
go get -u github.com/darkweak/go-esi
```

## Available as middleware
- [ ] Caddy
- [ ] Træfik
- [ ] Roadrunner

## TODO
- [x] choose tag
- [x] comment tag
- [ ] escape tag
- [x] include tag
- [x] remove tag
- [x] otherwise tag
- [ ] try tag
- [x] vars tag
- [x] when tag