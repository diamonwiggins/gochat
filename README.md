# sweetspotgg

### Project status: Alpha

Simple chat application writen in Go
Based on: https://github.com/scotch-io/go-realtime-chat

### Prerequisites

-[Docker]

-[Compose]

### Running the app
```
make install

If you want to detach eg. (-d) then run the following:
export DETACH=1 (or whatever it is for your shell)
```

### Additional make options
```
make clean
make build
make start
make stop
```

Listens on http://localhost:8080

[Docker]: https://docs.docker.com/install/
[Compose]: https://docs.docker.com/compose/install/