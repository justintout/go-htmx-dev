package gohtmxdev

import (
	"io"
	"log"
	"net/http"
	"time"
)

const DefaultEndpoint = "/_ghd/hotreload"

type Event string

const (
	pingEvent   Event = "event: ping"
	reloadEvent Event = "event: reload"
	closeEvent  Event = "event: close"
)

type Reloader struct {
	interval time.Duration
	logger   *log.Logger

	events         chan Event
	newClients     chan chan Event
	closingClients chan chan Event
	clients        map[chan Event]bool
}

func NewReloader(logger *log.Logger) (re *Reloader) {
	l := log.New(logger.Writer(), "reloader:: ", log.Lmsgprefix)
	re = &Reloader{
		interval:       10 * time.Second,
		logger:         l,
		events:         make(chan Event),
		newClients:     make(chan chan Event),
		closingClients: make(chan chan Event),
		clients:        make(map[chan Event]bool),
	}
	go re.listen()

	return re
}

func (re *Reloader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ec := make(chan Event)
	re.newClients <- ec

	defer func() {
		re.closingClients <- ec
	}()
	n := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-n
		re.closingClients <- ec
	}()
	ticker := time.NewTicker(re.interval)

	for {
		select {
		case <-ticker.C:
			io.WriteString(w, string(pingEvent))
			flusher.Flush()
		case evt := <-ec:
			io.WriteString(w, string(evt))
			flusher.Flush()
			if evt == closeEvent {
				return
			}
		}
	}
}

func (re *Reloader) Close() {
	for c := range re.clients {
		c <- closeEvent
	}
}

// TODO: pass path/filename so individual files can be
// reloaded - i.e. only reload specific path, instead of any page
func (re *Reloader) reload() {
	re.events <- reloadEvent
}

func (re *Reloader) listen() {
	re.logger.Printf("listening for events\n")
	for {
		select {
		case c := <-re.newClients:
			re.clients[c] = true
		case c := <-re.closingClients:
			delete(re.clients, c)
		case evt := <-re.events:
			re.logger.Printf("dispatching event: %s", evt)
			for c := range re.clients {
				c <- evt + "\n" + "data: \n\n"
			}
		}
	}
}
