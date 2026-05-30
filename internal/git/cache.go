package git

import (
	"os"
	"sync"
	"time"
)

type cacheEntry struct {
	status  RepoStatus
	modTime time.Time
}

var (
	cacheMu  sync.Mutex
	cacheMap = map[string]cacheEntry{}
)

// CachedStatus returns a cached RepoStatus if the .git directory mtime
// hasn't changed since the last check, otherwise recomputes and caches.
func CachedStatus(path string) (*RepoStatus, error) {
	fi, err := os.Stat(path + "/.git")
	if err != nil {
		return Status(path)
	}
	mt := fi.ModTime()

	cacheMu.Lock()
	entry, ok := cacheMap[path]
	if ok && entry.modTime.Equal(mt) {
		cacheMu.Unlock()
		s := entry.status
		return &s, nil
	}
	cacheMu.Unlock()

	s, err := Status(path)
	if err != nil {
		return nil, err
	}

	cacheMu.Lock()
	cacheMap[path] = cacheEntry{status: *s, modTime: mt}
	cacheMu.Unlock()

	return s, nil
}

// InvalidateCache removes a path from the cache, forcing a fresh status
// check on the next call.
func InvalidateCache(path string) {
	cacheMu.Lock()
	delete(cacheMap, path)
	cacheMu.Unlock()
}
