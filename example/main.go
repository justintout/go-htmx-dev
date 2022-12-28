package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	gohtmxdev "github.com/justintout/go-htmx-dev"
)

func main() {
	logger := log.New(os.Stdout, "", log.Lshortfile)

	reloader := gohtmxdev.NewReloader(logger)

	idxHandler, idxWatchErr, err := gohtmxdev.NewHandler[any](logger, reloader, "templates/index.gohtml", nil)
	if err != nil {
		panic(err)
	}
	go func() {
		for err := range idxWatchErr {
			logger.Println(err)
		}
	}()

	// TODO: better way to do this? need to synchronize prefix here, or
	// fix prefix stripping
	http.Handle("/_ghd/", gohtmxdev.Metahandler(reloader))
	http.Handle("/", idxHandler)

	addr := ":8008"
	fmt.Printf("serving on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("closing reloader connection(s)")
		reloader.Close()
		fmt.Printf("closing http server: %v\n", err)
		return
	}

}
