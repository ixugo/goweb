package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type EventMessage struct {
	id    string
	event string
	data  string
}

func (m *EventMessage) prepareMessage() []byte {
	var data bytes.Buffer
	if len(m.id) > 0 {
		data.WriteString(fmt.Sprintf("id: %s\n", strings.Replace(m.id, "\n", "", -1)))
	}
	if len(m.event) > 0 {
		data.WriteString(fmt.Sprintf("event: %s\n", strings.Replace(m.event, "\n", "", -1)))
	}
	if len(m.data) > 0 {
		lines := strings.Split(m.data, "\n")
		for _, line := range lines {
			data.WriteString(fmt.Sprintf("data: %s\n", line))
		}
	}
	data.WriteString("\n")
	return data.Bytes()
}

func NewEventMessage(event string, data map[string]any) *EventMessage {
	b, _ := json.Marshal(data)
	return &EventMessage{
		event: event,
		data:  string(b),
	}
}

func SendSSE(ch <-chan EventMessage, c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Content-Type", "text/event-stream")
	tick := time.NewTicker(40 * time.Millisecond)
	defer tick.Stop()
	var last *EventMessage
	var zero EventMessage
	for {
		select {
		case <-tick.C:
			if last != nil {
				_, _ = io.WriteString(c.Writer, fmt.Sprintf("%v\n", *last))
				c.Writer.Flush()
				last = nil
			}
		case v := <-ch:
			if v != zero {
				last = &v
				continue
			}
			if last != nil {
				_, _ = io.WriteString(c.Writer, fmt.Sprintf("%v\n", *last))
				c.Writer.Flush()
			}
			return
		}
	}
}

type Chunk struct {
	Total   int    `json:"total"`
	Current int    `json:"current"`
	Success int    `json:"success"`
	Failure int    `json:"failure"`
	Err     string `json:"err,omitempty"`
}

// SendChunkPro 高性能版
func SendChunkPro(ch <-chan Chunk, c *gin.Context) {
	if c == nil || c.Writer == nil {
		return
	}
	tick := time.NewTicker(40 * time.Millisecond)
	defer tick.Stop()
	var last *Chunk
	var zero Chunk
	var i int
	for {
		if i == 1 {
			c.Header("Cache-Control", "no-store")
			c.Header("Transfer-Encoding", "chunked")
			c.Header("Content-Type", "text/plain")
		}
		select {
		case <-tick.C:
			if last != nil {
				b, _ := json.Marshal(last)
				_, err := c.Writer.Write(append(b, '\n'))
				if err != nil {
					return
				}
				c.Writer.Flush()
				last = nil
			}
		case v := <-ch:
			i++
			if v != zero {
				last = &v
				continue
			}
			if last != nil {
				b, _ := json.Marshal(last)
				_, err := c.Writer.Write(append(b, '\n'))
				if err != nil {
					return
				}
				c.Writer.Flush()
			}
			return
		}
	}
}

// SendChunk 发送分块数据
func SendChunk(ch <-chan Chunk, c *gin.Context) {
	if c == nil || c.Writer == nil {
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Content-Type", "text/plain")
	var zero Chunk
	var i int
	for {
		i++
		v := <-ch
		if v == zero {
			return
		}
		b, _ := json.Marshal(v)
		_, err := c.Writer.Write(append(b, '\n'))
		if err != nil {
			return
		}
		c.Writer.Flush()
	}
}
