package cache

import (
	"sync"
	"time"
)

const MAX_CACHE_LEN uint32 = 1024

type ICache interface {
	Run()
}

type Caches struct {
	timeout time.Duration
	ccs     chan ICache
}

func (cs *Caches) Init(maxLen uint32, timeout time.Duration) {
	if maxLen == 0 {
		maxLen = MAX_CACHE_LEN
	}
	cs.timeout = timeout
	cs.ccs = make(chan ICache, maxLen)
	go func() {
		var c ICache
		for {
			c = <-cs.ccs
			c.Run()
		}
	}()
}

func (cs *Caches) Add(c ICache) bool {
	if cs.timeout == 0 {
		cs.ccs <- c
	} else {
		select {
		case cs.ccs <- c:
			return true
		case <-time.After(cs.timeout):
			return false
		}
	}
	return true
}

type CacheManager struct {
	lock      sync.RWMutex
	pCaches   *Caches
	mapCaches map[int32]*Caches
}

func (cm *CacheManager) Init() {
	cm.pCaches = &Caches{}
	cm.pCaches.Init(MAX_CACHE_LEN, 0)
	cm.mapCaches = make(map[int32]*Caches)
}

func (cm *CacheManager) AddCache(c ICache) bool {
	return cm.pCaches.Add(c)
}

func (cm *CacheManager) Set(t int32, maxLen uint32, timeout time.Duration) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	if _, ok := cm.mapCaches[t]; ok {
		return
	}
	cs := &Caches{}
	cs.Init(maxLen, timeout)
	cm.mapCaches[t] = cs
}

func (cm *CacheManager) AddCache2(t int32, c ICache) bool {
	var cs *Caches
	cm.lock.Lock()
	if cs, _ = cm.mapCaches[t]; cs == nil {
		cm.lock.Unlock()
		return false
	}
	cm.lock.Unlock()
	return cs.Add(c)
}

func GetCacheManager() *CacheManager {
	if g_cacheManager == nil {
		g_cacheManager = &CacheManager{}
		g_cacheManager.Init()
	}
	return g_cacheManager
}

var g_cacheManager *CacheManager
