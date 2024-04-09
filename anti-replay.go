package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strconv"
	"sync"
	"time"

	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	ep "github.com/wrossmorrow/envoy-extproc-sdk-go"
)

const (
	kSign  = "sign"
	kNonce = "nonce"

	kTimeSpan  = "timespan"
	kTimeStamp = "timestamp"
)

type antiReplayRequestProcessor struct {
	opts     *ep.ProcessingOptions
	timeSpan int64

	noncePool *ttlSet
}

func (s *antiReplayRequestProcessor) GetName() string {
	return "anti-replay"
}

func (s *antiReplayRequestProcessor) GetOptions() *ep.ProcessingOptions {
	return s.opts
}

func (s *antiReplayRequestProcessor) ProcessRequestHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func extract(m map[string]any, k string) string {
	vv, ok := m[k]
	if ok {
		return vv.(string)
	}
	return ""
}

func (s *antiReplayRequestProcessor) ProcessRequestBody(ctx *ep.RequestContext, body []byte) error {
	cancel := func(code int32) error {
		return ctx.CancelRequest(code, map[string]ep.HeaderValue{}, typev3.StatusCode_name[code])
	}

	var unstructure map[string]any

	err := json.Unmarshal(body, &unstructure)
	if err != nil {
		log.Printf("parse the request is failed: %v", err.Error())
		return cancel(400)
	}

	timestamp, _ := strconv.ParseInt(extract(unstructure, kTimeStamp), 10, 64)
	now := time.Now().Unix()
	if timestamp < now-s.timeSpan || timestamp > now+s.timeSpan {
		return cancel(403)
	}

	nonce := extract(unstructure, kNonce)
	if s.noncePool.exists(nonce) {
		return cancel(403)
	}
	s.noncePool.put(nonce)

	var (
		keys []string
		m    = map[string]string{}
		sign string
	)
	for k, v := range unstructure {
		val := v.(string)
		if len(val) != 0 {
			if k != kSign {
				keys = append(keys, k)
				m[k] = val
			} else {
				sign = val
			}
		}
	}

	slices.Sort(keys)

	buf := &bytes.Buffer{}
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("%s=%s&", k, m[k]))
	}

	buf.Truncate(buf.Len() - 1)

	raw := buf.Bytes()
	fmt.Println(string(raw))

	hash := md5.Sum(raw)

	md5Hex := hex.EncodeToString(hash[:])
	if sign != md5Hex {
		return cancel(403)
	}

	return ctx.ContinueRequest()
}

func (s *antiReplayRequestProcessor) ProcessRequestTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (s *antiReplayRequestProcessor) ProcessResponseHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (s *antiReplayRequestProcessor) ProcessResponseBody(ctx *ep.RequestContext, body []byte) error {
	return ctx.ContinueRequest()
}

func (s *antiReplayRequestProcessor) ProcessResponseTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (s *antiReplayRequestProcessor) Init(opts *ep.ProcessingOptions, nonFlagArgs []string) error {
	s.opts = opts
	s.timeSpan = 15 * 60

	var i int
	nArgs := len(nonFlagArgs)
	for ; i < nArgs-1; i++ {
		if nonFlagArgs[i] == kTimeSpan {
			break
		}
	}

	if i == nArgs {
		log.Printf("the argument: 'timespan' is missing, use the default.\n")
	} else {
		timeSpan, _ := strconv.ParseInt(nonFlagArgs[i+1], 10, 64)
		if timeSpan == 0 {
			log.Printf("parse the value for parameter: 'timespan' is failed,use the default.\n")
		} else {
			s.timeSpan = timeSpan
			log.Printf("the timespan is: %d.\n", s.timeSpan)
		}
	}

	s.noncePool = newTTLSet(s.timeSpan)
	go s.noncePool.evictExpired()

	return nil
}

func (s *antiReplayRequestProcessor) Finish() {
	s.noncePool.finish()
}

type ttlSet struct {
	mu       sync.Mutex
	pool     map[string]int64
	timeSpan int64
	chEvict  chan struct{}
	done     chan struct{}
}

func newTTLSet(timeSpan int64) *ttlSet {
	return &ttlSet{
		pool:     map[string]int64{},
		chEvict:  make(chan struct{}),
		done:     make(chan struct{}),
		timeSpan: timeSpan,
	}
}

func (c *ttlSet) put(v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.existsLocked(v) {
		c.pool[v] = time.Now().Unix()
	}
}

func (c *ttlSet) evictExpired() {
	defer close(c.done)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.chEvict:
			return

		case <-ticker.C:
			now := time.Now().Unix()
			c.mu.Lock()
			for k, v := range c.pool {
				if v < now-c.timeSpan {
					delete(c.pool, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *ttlSet) existsLocked(v string) bool {
	_, ok := c.pool[v]
	return ok
}

func (c *ttlSet) exists(v string) bool {
	if len(v) == 0 {
		return true
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.existsLocked(v)
}

func (c *ttlSet) finish() {
	close(c.chEvict)
	<-c.done
}
