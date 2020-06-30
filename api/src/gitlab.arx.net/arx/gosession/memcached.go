package gosession

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
)

// MemcachedProvider stores the session in a memcached server
type MemcachedProvider struct {
	Connection *memcache.Client
	KeyPrefix  string
}

// Save the session in memcache
func (provider *MemcachedProvider) Save(
	id string, data map[string]interface{}) error {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(data)

	if err != nil {
		return err
	}

	err = provider.Connection.Set(&memcache.Item{
		Key:   fmt.Sprintf("%s-%s", provider.KeyPrefix, id),
		Value: buffer.Bytes()})

	return err
}

// Delete the session with that id
func (provider *MemcachedProvider) Delete(id string) error {
	return provider.Connection.Delete(fmt.Sprintf("%s-%s", provider.KeyPrefix, id))
}

// Get session data
func (provider *MemcachedProvider) Get(id string,
	data *map[string]interface{}) (bool, error) {

	item, err := provider.Connection.Get(fmt.Sprintf("%s-%s", provider.KeyPrefix, id))

	if err == memcache.ErrCacheMiss {
		return false, nil
	} else if err != nil {
		return false, err
	}

	var buffer bytes.Buffer

	buffer.Write(item.Value)

	decoder := gob.NewDecoder(&buffer)

	err = decoder.Decode(data)

	if err != nil {
		return false, err
	}

	return true, nil
}
