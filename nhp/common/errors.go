package common

import (
	"strconv"
	"strings"
)

var errorMap map[string]*Error = make(map[string]*Error)

var ErrorMsgLanguageLocale string = "EN"

type Error struct {
	code  string
	msgEN string
	msgCH string
}

// implment NhpError interface
func (e *Error) Error() string {
	switch strings.ToUpper(ErrorMsgLanguageLocale) {
	case "EN", "EN-US", "EN-GB":
		return e.msgEN
	case "ZH", "ZH-HANS":
		return e.msgCH
	}
	return ""
}

func (e *Error) ErrorCode() string {
	return e.code
}

func (e *Error) ErrorNumber() int {
	n, _ := strconv.Atoi(e.code)
	return n
}

func newError(code string, enStr string, chStr string) *Error {
	e := &Error{
		code:  code,
		msgEN: enStr,
		msgCH: chStr,
	}
	errorMap[code] = e
	return e
}

func ErrorToErrorCode(err error) string {
	e, ok := err.(*Error)
	if ok {
		return e.ErrorCode()
	}
	return ""
}

func ErrorToString(err error) string {
	e, ok := err.(*Error)
	if ok {
		return e.Error()
	}
	return ""
}

func ErrorCodeToError(code string) *Error {
	e, found := errorMap[code]
	if found {
		return e
	}
	return nil // should not happen
}

// application errors
var (
	// generic
	ErrSuccess                             = newError("0", "", "")
	ErrExit                                = newError("1", "must exit", "立即退出")
	ErrJsonParseFailed                     = newError("50001", "json parse failed", "json解析失败")
	ErrTransactionIdNotFound               = newError("50002", "transaction id not found", "无法找到交互id")
	ErrTransactionFailedByTimeout          = newError("50003", "transaction failed due to time out", "请求超时，交互失败")
	ErrTransactionFailedByClosedConnection = newError("50004", "transaction failed by closed connection", "由于连接中断，交互失败")
	ErrTransactionFailedByClosedDevice     = newError("50005", "transaction failed by closed device", "由于设备停止，交互失败")
	ErrTransactionRepliedWithWrongType     = newError("50006", "transaction replied wrong type", "交互回应了错误的消息类型")
	ErrPacketToMessageRoutineStopped       = newError("50007", "packet to message routine stopped", "消息处理线程已停止")
	ErrInvalidIpAddress                    = newError("50008", "invalid ip address", "ip地址无效")
	ErrPacketEncryptionFailed              = newError("50009", "packet encryption failed", "报文加密失败")

	// agent
	ErrKnockUserNotSpecified   = newError("51001", "knock user not specified", "没有指定敲门用户")
	ErrKnockServerNotFound     = newError("51002", "failed to find knock server", "无法找到敲门服务器")
	ErrKnockTerminatedByCookie = newError("51003", "knock terminated by cookie", "敲门被cookie包中止")

	// agentsdk
	ErrNoAgentInstance = newError("51100", "agent instance does not exist", "未创建agent实例")
	ErrInvalidInput    = newError("51101", "invalid input parameter", "无效的输入参数")

	// server
	ErrKnockApiRequestFailed       = newError("52001", "knock api request failed", "敲门api请求失败")
	ErrAuthServiceProviderNotFound = newError("52002", "failed to find auth service provider", "无法找到服务提供商")
	ErrACConnectionNotFound        = newError("52003", "failed to find ac connection", "无法找到门禁连接")
	ErrResourceNotFound            = newError("52004", "failed to find resource", "无法找到资源")
	ErrServerACOpsFailed           = newError("52005", "server ac operation failed", "服务器请求门禁操作失败")
	ErrAuthHandlerNotFound         = newError("52006", "failed to find auth handler", "无法找到验证处理接口")
	ErrBackendAuthRequired         = newError("52007", "server backend auth required", "服务器需要后端敲门验证")
	ErrUrlPathInvalid              = newError("52008", "client request url path is invalid", "请求路径无效")

	// ac
	ErrACOperationFailed       = newError("53001", "ac operation failed", "门禁操作失败")
	ErrACEmptyPassAddress      = newError("53002", "pass address is empty", "放行地址为空")
	ErrACIPSetNotFound         = newError("53003", "ipset not found", "无法找到ipset")
	ErrACIPSetOperationFailed  = newError("53004", "ipset operation failed", "ipset操作失败")
	ErrACTempPortListenFailed  = newError("53005", "temporary port listening failed", "临时端口监听失败")
	ErrACResolveTempPortFailed = newError("53006", "resolve temparory port failed", "解析临时端口失败")

	// api
	ErrHttpRequestFailed           = newError("54001", "http request failed", "http请求失败")
	ErrHttpResponseFormatError     = newError("54002", "http response format error", "http响应格式错误")
	ErrHttpReturnedWithError       = newError("54003", "http returns with error", "http返回带有错误")
	ErrHttpResourceAddressNotFound = newError("54004", "http resource address not found", "http无法找到资源地址")
)
