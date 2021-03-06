package xredis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
)

// Helper is an extensions of redis.Conn.
type Helper struct {
	conn redis.Conn
}

// WithDB creates a Helper with redis.Conn.
func WithConn(conn redis.Conn) *Helper {
	return &Helper{conn: conn}
}

// GetDB gets the original redis.Conn.
func (h *Helper) GetConn() redis.Conn {
	return h.conn
}

// DeleteAll: KEYS xxx -> DEL xxx
func (h *Helper) DeleteAll(pattern string) (total int, del int, err error) {
	keys, err := redis.Strings(h.conn.Do("KEYS", pattern))
	if err != nil {
		return 0, 0, err
	}

	cnt := 0
	var someErr error
	for _, key := range keys {
		result, err := redis.Int(h.conn.Do("DEL", key))
		if err == nil {
			cnt += result
		} else if someErr == nil {
			someErr = err
		}
	}

	return len(keys), cnt, someErr
}

// SetAll: SET xxx
func (h *Helper) SetAll(keys []string, values []string) (total int, add int, err error) {
	cnt := 0
	if len(keys) != len(values) {
		return 0, 0, fmt.Errorf("the length of keys and values is different")
	}

	var someErr error
	for idx := range keys {
		key := keys[idx]
		value := values[idx]

		_, err := h.conn.Do("SET", key, value)
		if err == nil {
			cnt++
		} else if someErr == nil {
			someErr = err
		}
	}

	return len(keys), cnt, someErr
}

// SetExAll: SET xxx yyy
func (h *Helper) SetExAll(keys []string, values []string, exs []int64) (total int, add int, err error) {
	cnt := 0
	if len(keys) != len(values) && len(keys) != len(exs) {
		return 0, 0, fmt.Errorf("the length of keys, values and exs is different")
	}

	var someErr error
	for idx := range keys {
		key := keys[idx]
		value := values[idx]
		ex := exs[idx]

		_, err := h.conn.Do("SET", key, value, ex)
		if err == nil {
			cnt++
		} else if someErr == nil {
			someErr = err
		}
	}

	return len(keys), cnt, someErr
}
