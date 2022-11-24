package middlewares

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"code.jshyjdtech.com/godev/hykit/log"

	"github.com/gin-gonic/gin"
)

func ReporterParams(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		beg := time.Now()
		ctx := c.Request.Context()

		defer func() {
			if rec := recover(); rec != nil {
				logger.Errorc(ctx, "AccessLoggerPanic, err: %v stack: %v", rec, string(debug.Stack()))

				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}()

		// 写入response Body之前把body拷贝一份到buffer中
		writer := &bodyLogWriter{
			ResponseWriter: c.Writer,
			bodyBuf:        bytes.NewBufferString(""),
		}
		c.Writer = writer
		reqBuf, _ := c.GetRawData()

		logger.Infoc(ctx, "入参: method[%v],size[%v],host[%v],path[%v],ip[%v],header[%v],body:[%s]",
			c.Request.Method, len(reqBuf), c.Request.Host, c.Request.URL.Path,
			c.ClientIP(), c.Request.Header, string(reqBuf))

		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBuf))

		c.Next()

		debugBuffer := writer.bodyBuf.String()
		logger.Infoc(ctx, "出参: path[%v],Body:[%v],statusCode[%v],size[%v],cost[%v]",
			c.Request.URL.Path, debugBuffer,
			c.Writer.Status(), c.Writer.Size(),
			time.Since(beg).Seconds())
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	bodyBuf *bytes.Buffer
}

func (w bodyLogWriter) Write(buf []byte) (int, error) {
	w.bodyBuf.Write(buf)
	return w.ResponseWriter.Write(buf)
}

func ReportHeaderSet(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		rows := strings.Split(c.Request.URL.Path, "/")
		fileName := rows[len(rows)-1]
		c.Writer.Header().Set("content-type", "application/octet-stream")
		c.Writer.Header().Set("content-disposition", fmt.Sprintf("attachment;filename=%s", fileName))

		c.Next()
		logger.Debugc(ctx, "出参Header: "+
			"content-type[%s],  content-disposition[%s]", c.Writer.Header().Get("content-type"), c.Writer.Header().Get("content-disposition"))
	}
}
