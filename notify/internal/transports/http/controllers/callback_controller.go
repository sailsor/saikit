package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"notify/internal/application"
	"notify/internal/infra"
	"notify/internal/transports/http/dto"
	"notify/pkg/response"
	"strings"
)

type CallbackController struct {
	*infra.Infra
	UnApp *application.UnionJsCallbackSvc
}

// UnionJSCallback 江苏银联异步回调
func (cb *CallbackController) UnionJSCallback(c *gin.Context) {
	logger := cb.Logger
	var (
		err error
		req = new(dto.UnionNotify)
		ctx = c.Request.Context()
	)

	encoding := c.PostForm("encoding")
	logger.Infof("encoding: [%s]", encoding)

	err = c.Request.ParseForm()
	if err != nil {
		logger.Errorc(ctx, "UnionJSCallback:ParseForm 解析参数失败[%s]", err)
		response.FailWithMessage(c, fmt.Sprintf("ParseForm error:[%s]", err))
		return
	}

	var postForm strings.Builder
	postForm.WriteString("打印Form域:\n")

	for k, v := range c.Request.PostForm {
		postForm.WriteString(fmt.Sprintf("Field[%s]=[%s]\n", k, v[0]))
	}
	logger.Infoc(ctx, postForm.String())

	err = c.ShouldBind(req)
	if err != nil {
		logger.Errorf("ShouldBind失败[%s]", err)
	}

	cb.UnApp.UnNotify = req
	err = cb.UnApp.VerifySign(ctx)
	if err != nil {
		logger.Errorf("%s", err)
	}

	c.String(http.StatusOK, "SUCCESS")
	return
}

func (cb *CallbackController) Notify(c *gin.Context) {
	logger := cb.Logger
	ctx := context.Background()

	logger.Infoc(ctx, "响应body为空")

	c.String(http.StatusOK, "")
	return
}

func (cb *CallbackController) NotifyWithSUCCESS(c *gin.Context) {
	logger := cb.Logger
	ctx := context.Background()

	msg := "SUCCESS"
	logger.Infoc(ctx, "响应body[%s]", msg)

	c.String(http.StatusOK, msg)
	return
}

func (cb *CallbackController) NotifyWithOK(c *gin.Context) {
	logger := cb.Logger
	ctx := context.Background()

	msg := "OK"
	logger.Infoc(ctx, "响应body[%s]", msg)

	c.String(http.StatusOK, msg)
	return
}

func (cb *CallbackController) NotifyWith502(c *gin.Context) {
	logger := cb.Logger
	ctx := context.Background()

	logger.Infoc(ctx, "响应502")

	c.Writer.WriteHeader(http.StatusBadGateway)
	return
}
