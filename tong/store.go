package tong

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/gocolly/colly/v2"
	whatwgUrl "github.com/nlnwa/whatwg-url/url"
	log "github.com/sirupsen/logrus"
)

type Store interface {
	// Init initializes the storage
	Init() error
	// AddRequest adds a serialized request to the queue
	AddRequest([]byte) error
	// GetRequest pops the next request from the queue
	// or returns error if the queue is empty
	GetRequest() ([]byte, error)
	// QueueSize returns with the size of the queue
	QueueSize() (int, error)
	// Visited receives and stores a request ID that is visited by the Collector
	Visited(requestID uint64) error
	// IsVisited returns true if the request was visited before IsVisited
	// is called
	IsVisited(requestID uint64) (bool, error)
	// Cookies retrieves stored cookies for a given host
	Cookies(u *url.URL) string
	// SetCookies stores cookies for a given host
	SetCookies(u *url.URL, cookies string)
}

type BloomStore struct {
	Client    *redis.Client
	Id        string
	TongsName string
	Expires   time.Duration
	mu        sync.RWMutex
	IsQueue   bool //是否是队列模式
}

// Init initializes the redis storage
func (s *BloomStore) Init() error {
	if s.Client == nil {
		return errors.New(fmt.Sprintf("【%s】未设置存储器", s.Id))
	}
	return nil
}

// Clear removes all entries from the storage
func (s *BloomStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.Client.Keys(fmt.Sprintf("%s:cookie", "*"))
	keys, err := r.Result()
	if err != nil {
		return err
	}
	r2 := s.Client.Keys(s.Id + ":request:*")
	keys2, err := r2.Result()
	if err != nil {
		return err
	}
	keys = append(keys, keys2...)
	keys = append(keys, s.getQueueID())
	return s.Client.Del(keys...).Err()
}

// Visited 非队列调用时通过该方法判断去重
func (s *BloomStore) Visited(requestID uint64) error {
	return s.Client.Do("BF.ADD", s.getBloomID(), requestID).Err()
}

// IsVisited 非队列调用时通过该方法判断去重
func (s *BloomStore) IsVisited(requestID uint64) (bool, error) {
	_, err := s.Client.Do("BF.EXISTS", s.getBloomID(), requestID).Bool()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// SetCookies implements colly/storage..SetCookies()
func (s *BloomStore) SetCookies(u *url.URL, cookies string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.Client.HSet(s.getCookieID(), u.Host, cookies).Err()
	if err != nil {
		// return nil
		log.Printf("SetCookies() .Set error %s", err)
		return
	}
}

// Cookies implements colly/storage.Cookies()
func (s *BloomStore) Cookies(u *url.URL) string {
	// TODO(js) Cookie methods currently have no way to return an error.

	s.mu.RLock()
	cookiesStr, err := s.Client.HGet(s.getCookieID(), u.Host).Result()
	s.mu.RUnlock()
	if err == redis.Nil {
		cookiesStr = ""
	} else if err != nil {
		// return nil, err
		log.Printf("Cookies() .Get error %s", err)
		return ""
	}
	return cookiesStr
}

// AddRequest implements queue.Storage.AddRequest() function
func (s *BloomStore) AddRequest(r []byte) error {
	var req colly.Request
	json.Unmarshal(r, &req)
	url := req.URL.String()
	if req.Method == "GET" {
		exists, err := s.Client.Do("BF.EXISTS", s.getBloomID(), url).Bool()
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
	}
	_, err := s.Client.TxPipelined(func(pipe redis.Pipeliner) error {
		pipe.RPush(s.getQueueID(), r)
		s.Client.Do("BF.ADD", s.getBloomID(), url)
		return nil
	})
	return err
}

// GetRequest implements queue.Storage.GetRequest() function
func (s *BloomStore) GetRequest() ([]byte, error) {
	r, err := s.Client.BLPop(10*time.Minute, s.getQueueID()).Result()
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errors.New("queue is empty")
	}
	return []byte(r[1]), err
}

// QueueSize implements queue.Storage.QueueSize() function
func (s *BloomStore) QueueSize() (int, error) {
	i, err := s.Client.LLen(s.getQueueID()).Result()
	return int(i), err
}

func (s *BloomStore) getIDStr(ID uint64) string {
	return fmt.Sprintf("%s:request:%d", s.Id, ID)
}

func (s *BloomStore) getCookieID() string {
	return fmt.Sprintf("%s:cookie", s.TongsName)
}

func (s *BloomStore) getQueueID() string {
	return fmt.Sprintf("%s:queue", s.Id)
}
func (s *BloomStore) getBloomID() string {
	if !Config.Bloom.Open || Config.Bloom.Alone {
		return s.Id
	}
	if !Config.Bloom.Alone {
		return s.TongsName
	}
	return s.Id
}

type TongsStore struct {
	Client    *redis.Client
	Id        string
	TongsName string
	Expires   time.Duration
	mu        sync.RWMutex
	IsQueue   bool //是否是队列模式
}

// Init initializes the redis storage
func (s *TongsStore) Init() error {
	if s.Client == nil {
		return errors.New(fmt.Sprintf("【%s】未设置存储器", s.Id))
	}
	return nil
}

// Clear removes all entries from the storage
func (s *TongsStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.Client.Keys(fmt.Sprintf("%s:cookie", "*"))
	keys, err := r.Result()
	if err != nil {
		return err
	}
	r2 := s.Client.Keys(s.Id + ":request:*")
	keys2, err := r2.Result()
	if err != nil {
		return err
	}
	keys = append(keys, keys2...)
	keys = append(keys, s.getQueueID())
	return s.Client.Del(keys...).Err()
}

// Visited 非队列调用时通过该方法判断去重
func (s *TongsStore) Visited(requestID uint64) error {
	return s.Client.SAdd(s.getVisitedID(), requestID).Err()
}

// IsVisited 非队列调用时通过该方法判断去重
func (s *TongsStore) IsVisited(requestID uint64) (bool, error) {
	visited, err := s.Client.SIsMember(s.getVisitedID(), requestID).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return visited, nil
}

// SetCookies implements colly/storage..SetCookies()
func (s *TongsStore) SetCookies(u *url.URL, cookies string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.Client.HSet(s.getCookieID(), u.Host, cookies).Err()
	if err != nil {
		// return nil
		log.Printf("SetCookies() .Set error %s", err)
		return
	}
}

// Cookies implements colly/storage.Cookies()
func (s *TongsStore) Cookies(u *url.URL) string {
	// TODO(js) Cookie methods currently have no way to return an error.

	s.mu.RLock()
	cookiesStr, err := s.Client.HGet(s.getCookieID(), u.Host).Result()
	s.mu.RUnlock()
	if err == redis.Nil {
		cookiesStr = ""
	} else if err != nil {
		// return nil, err
		log.Printf("Cookies() .Get error %s", err)
		return ""
	}
	return cookiesStr
}

// AddRequest implements queue.Storage.AddRequest() function
func (s *TongsStore) AddRequest(r []byte) error {
	var req map[string]interface{}
	json.Unmarshal(r, &req)
	var reqId uint64
	if req["Method"] == "GET" {
		reqId = requestHash(req["URL"].(string), nil)
		visited, err := s.IsVisited(reqId)
		if err != nil {
			return err
		}
		if visited {
			return nil
		}
	}

	_, err := s.Client.TxPipelined(func(pipe redis.Pipeliner) error {
		pipe.RPush(s.getQueueID(), r)
		if reqId != 0 {
			pipe.SAdd(s.getVisitedID(), reqId)
		}
		return nil
	})
	return err
}

// GetRequest implements queue.Storage.GetRequest() function
func (s *TongsStore) GetRequest() ([]byte, error) {
	r, err := s.Client.BLPop(10*time.Minute, s.getQueueID()).Result()
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errors.New("queue is empty")
	}
	return []byte(r[1]), err
}

// QueueSize implements queue.Storage.QueueSize() function
func (s *TongsStore) QueueSize() (int, error) {
	i, err := s.Client.LLen(s.getQueueID()).Result()

	return int(i), err
}

func (s *TongsStore) getIDStr(ID uint64) string {
	return fmt.Sprintf("%s:request:%d", s.Id, ID)
}

func (s *TongsStore) getCookieID() string {
	return fmt.Sprintf("%s:cookie", s.TongsName)
}

func (s *TongsStore) getQueueID() string {
	return fmt.Sprintf("%s:queue", s.Id)
}
func (s *TongsStore) getVisitedID() string {
	if Config.Bloom.Alone {
		return fmt.Sprintf("%s:visited", s.Id)
	} else {
		return fmt.Sprintf("%s:visited", s.TongsName)
	}
}

func requestHash(url string, body io.Reader) uint64 {
	h := fnv.New64a()
	// reparse the url to fix ambiguities such as
	// "http://example.com" vs "http://example.com/"
	parsedWhatwgURL, err := whatwgUrl.Parse(url)
	if err == nil {
		h.Write([]byte(parsedWhatwgURL.String()))
	} else {
		h.Write([]byte(url))
	}
	if body != nil {
		io.Copy(h, body)
	}
	return h.Sum64()
}
