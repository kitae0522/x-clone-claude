package worker

import (
	"sync"

	"github.com/kitae0522/twitter-clone-claude/media-service/internal/model"
)

// Registry holds in-memory media metadata and processing status.
type Registry struct {
	mu    sync.RWMutex
	items map[string]*model.Media
}

func NewRegistry() *Registry {
	return &Registry{
		items: make(map[string]*model.Media),
	}
}

func (r *Registry) Set(id string, media *model.Media) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[id] = media
}

func (r *Registry) Get(id string) (*model.Media, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.items[id]
	return m, ok
}

func (r *Registry) UpdateStatus(id string, status model.Status, errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.items[id]; ok {
		m.Status = status
		m.Error = errMsg
	}
}

func (r *Registry) UpdateDimensions(id string, width, height int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.items[id]; ok {
		m.Width = width
		m.Height = height
	}
}

func (r *Registry) Delete(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
}
