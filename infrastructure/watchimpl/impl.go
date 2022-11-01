package watchimpl

type Watcher struct{}

func NewWatcher(cfg *Config) *Watcher {
	return &Watcher{}
}

func (w *Watcher) Run() {}

func (w *Watcher) Exit() {}
