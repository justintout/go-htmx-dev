package gohtmxdev

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
)

type Data[D any] struct {
	HotReload bool
	Data      D
}

type Handler[D any] struct {
	tc       chan *template.Template
	w        *watcher.Watcher
	interval time.Duration
	logger   *log.Logger
	reloader *Reloader

	mu   *sync.Mutex
	fn   string
	tmpl *template.Template
	data Data[D]
}

func NewHandler[D any](logger *log.Logger, reloader *Reloader, fn string, data *D) (*Handler[D], chan error, error) {
	l := log.New(logger.Writer(), fn+":: ", log.Lmsgprefix)
	tmpl, err := template.ParseFiles(fn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse %s: %w", fn, err)
	}
	w := watcher.New()
	w.SetMaxEvents(1)
	if err := w.Add(fn); err != nil {
		return nil, nil, fmt.Errorf("failed to add watcher for %s: %w", fn, err)
	}
	h := &Handler[D]{
		tc:       make(chan *template.Template),
		w:        w,
		interval: 100 * time.Millisecond,
		logger:   l,
		reloader: reloader,
		mu:       new(sync.Mutex),
		fn:       fn,
		tmpl:     tmpl,
		data:     Data[D]{HotReload: true},
	}
	// TODO: use error channel?
	ec := h.watch()
	return h, ec, nil
}

func (h *Handler[D]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.tmpl.Execute(w, h.data)
}

func (h *Handler[D]) watch() chan error {
	go func() {
		for {
			select {
			case evt := <-h.w.Event:
				h.logger.Printf("saw file event: %s\n", evt.Op)
				if evt.Op == watcher.Rename {
					h.logger.Printf("file renamed from %s to %s\n", evt.OldPath, evt.Path)
					h.mu.Lock()
					h.fn = evt.Path
					h.mu.Unlock()
				}
				n, err := template.ParseFiles(h.fn)
				if err != nil {
					h.logger.Printf("failed to parse %s: %v\n", h.fn, err)
					continue
				}
				h.mu.Lock()
				h.tmpl = n
				h.mu.Unlock()
				h.triggerReload()
			case err := <-h.w.Error:
				h.logger.Printf("error watching %s: %v\n", h.fn, err)
			case <-h.w.Closed:
				h.logger.Printf("watcher for %s closed\n", h.fn)
				return
			}
		}
	}()
	c := make(chan error)
	go func() {
		c <- h.w.Start(h.interval)
	}()
	return c
}

func (h *Handler[D]) Update(data D) {
	h.mu.Lock()
	h.data.Data = data
	h.mu.Unlock()
	h.triggerReload()
}

func (h *Handler[D]) triggerReload() {
	h.reloader.reload()
}
