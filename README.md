# go-htmx-dev

> Toolbox for using [htmx](https://htmx.org) in [Go](https://go.dev).

This is an early-stage work-in-progress.  

`go-htmx-dev` provides hot-reload for [Go HTML templates](https://pkg.go.dev/html/template), with the eventual goal of providing an SSR-based hot-reload mechanism for HTMX websites.  

Right now, the "hot reload" part kinda works. Still **many** design decisions to be made. 

## References and Dependencies 
- [radovskyb/watcher](https://github.com/radovskyb/watcher): cross-platform file watcher, to trigger hot reloading
- [Writing a Server Sent Events server in Go - Ismael Celis](https://thoughtbot.com/blog/writing-a-server-sent-events-server-in-go): Basis for the SSE handler