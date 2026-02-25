package common_response_format

import (
	example0 "VideoWeb/biz/model/common/example"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

type CommonResponse struct {
	Code    example0.ErrorCode `thrift:"Code,1,default,ErrorCode" form:"Code" json:"Code" query:"Code"`
	Message string             `thrift:"Message,2" form:"Message" json:"Message" query:"Message"`
	Data    interface{}        `thrift:"Data,3" form:"Data" json:"Data" query:"Data"`
}

func Success(c *app.RequestContext, msg string, data interface{}) {
	c.JSON(http.StatusOK, CommonResponse{
		Code:    example0.ErrorCode_SUCCESS,
		Message: msg + " success!",
		Data:    data,
	})
}

func Fail(c *app.RequestContext, httpStatus int, code example0.ErrorCode, msg string) {
	c.JSON(httpStatus, CommonResponse{
		Code:    code,
		Message: msg,
		Data:    nil,
	})
}
