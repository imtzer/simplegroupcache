// @Description: group
// @Author: tao
// @Date: 2021/10/08 19:14
package group

import (
	"errors"
	"fmt"
	"simplecache/lru"
	"sync"
)

// 缓存未命中时的回调方法，由用户来提供实现
type Getter interface {
	Get(key string) ([]byte, error)
}

// 接口型函数, 调用接口的方法相当于调用自己，让一个函数能够当成回调传入
type GetFunc func(key string) ([]byte, error)

func (g GetFunc) Get(key string) ([]byte, error) {
	return g(key)
}

// 一种类型的缓存
type Member struct {
	name   string
	getter Getter
	mCache *lru.Cache
}

type Group struct {
	mu sync.RWMutex
	group map[string]*Member
}

// 提供一个默认的group
var defaultGroup Group = Group{
	group: make(map[string]*Member),
}

func NewMember(name string, cacheSize int64, getter Getter, onEvicted func(key string, value lru.Value)) *Member {
	if getter == nil {
		panic("getter is nil")
	}
	defaultGroup.mu.Lock()
	defer defaultGroup.mu.Unlock()
	m := &Member{
		name: name,
		getter: getter,
		mCache: lru.NewCache(cacheSize, onEvicted),
	}
	defaultGroup.group[name] = m
	return m
}

func GetMember(name string) *Member {
	defaultGroup.mu.RLock()
	m := defaultGroup.group[name]
	defaultGroup.mu.RUnlock()
	return m
}


// Group的方法

func (g *Group) NewMember(name string, cacheSize int64, getter Getter, onEvicted func(key string, value lru.Value)) *Member {
	if getter == nil {
		panic("getter is nil")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	m := &Member{
		name: name,
		getter: getter,
		mCache: lru.NewCache(cacheSize, onEvicted),
	}
	g.group[name] = m
	return m
}

func (g *Group) GetMember(name string) *Member {
	g.mu.RLock()
	m := g.group[name]
	g.mu.RUnlock()
	return m
}


// Member的方法

func (m *Member) Get(key string) (lru.ByteValue, error) {
	if len(key) == 0 {
		return lru.ByteValue{}, errors.New("invalid key")
	}
	if v, err := m.mCache.Get(key); err == nil {
		fmt.Println("cache hit")
		return v, nil
	}
	return m.load(key)
}

func (m *Member) load(key string) (lru.ByteValue, error) {
	return m.getLocally(key)
}

func (m *Member) getLocally(key string) (lru.ByteValue, error) {
	 bytes, err := m.getter.Get(key)
	 if err != nil {
		return lru.ByteValue{}, err
	}

	value := lru.ByteValue{B: bytes}
	m.mCache.Add(key, value)
	return value, nil
}