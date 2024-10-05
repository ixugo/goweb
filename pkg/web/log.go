package web

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	max := len(b)
	if max > 100 {
		max = 100
	}
	w.body.Write(b[:max])
	return w.ResponseWriter.Write(b)
}

// Logger ..
func Logger(log *slog.Logger, recordBody bool, ignore func(path string) bool, fileds ...func(*gin.Context) []any) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 过滤无须记录日志的
		path := c.Request.URL.Path
		if ignore(path) {
			c.Next()
			return
		}
		var reqBody string
		var blw bodyLogWriter
		if recordBody {
			// 请求参数
			raw, err := c.GetRawData()
			if err != nil {
				slog.Error("logger", "err", err)
			}
			maxL := len(raw)
			if maxL > 100 {
				maxL = 100
			}
			reqBody = string(raw[:maxL])

			c.Request.Body = io.NopCloser(bytes.NewReader(raw))
			// 响应参数
			blw = bodyLogWriter{
				body:           bytes.NewBuffer(nil),
				ResponseWriter: c.Writer,
			}
			c.Writer = &blw
		}

		now := time.Now()
		// traceid := trace.SpanContextFromContext(c.Request.Context()).TraceID().String()
		SetTraceID(c, uuid.NewString())
		c.Next()

		code := c.Writer.Status()
		out := []any{
			"uid", uid,
			"method", c.Request.Method,
			"path", path,
			"query", c.Request.URL.RawQuery,
			"remoteaddr", c.ClientIP(),
			"statuscode", code,
			"since", time.Since(now).Milliseconds(),
			"trace_id", MustTraceID(c),
		}
		if recordBody {
			out = append(out, []any{"request_body", reqBody, "response_body", blw.body.String()}...)
		}
		for _, v := range fileds {
			out = append(out, v(c))
		}
		if code >= 200 && code < 400 {
			log.Info("OK", out...)
			return
		}
		// 约定: 返回给客户端的错误，记录的 key 为 responseErr
		errStr, _ := c.Get("responseErr")
		if !(code == 404 || code == 401) {
			out = append(out, []any{"err", errStr})
		}
		log.Warn("Bad", out...)
	}
}
