package plugins

import (
	"fmt"
	"path/filepath"
	"plugin"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/log"
	"github.com/OpenNHP/opennhp/utils"
	"github.com/gin-gonic/gin"
)

var ExeDirPath string

type PluginHandler interface {
	Version() string
	Signature() string
	ExportedData() *PluginParamsOut
	Init(*PluginParamsIn) error
	Close() error
	RequestOTP(*common.NhpOTPRequest, *NhpServerPluginHelper) error
	RegisterAgent(*common.NhpRegisterRequest, *NhpServerPluginHelper) (*common.ServerRegisterAckMsg, error)
	ListService(*common.NhpListRequest, *NhpServerPluginHelper) (*common.ServerListResultMsg, error)
	AuthWithNHP(*common.NhpAuthRequest, *NhpServerPluginHelper) (*common.ServerKnockAckMsg, error)
	AuthWithHttp(*gin.Context, *common.HttpKnockRequest, *HttpServerPluginHelper) (*common.ServerKnockAckMsg, error)
}

type PluginHandlerSymbol struct {
	sVersion       plugin.Symbol
	sSignature     plugin.Symbol
	sExportedData  plugin.Symbol
	sInit          plugin.Symbol
	sClose         plugin.Symbol
	sRequestOTP    plugin.Symbol
	sRegisterAgent plugin.Symbol
	sListService   plugin.Symbol
	sAuthWithNHP   plugin.Symbol
	sAuthWithHttp  plugin.Symbol
}

var errPluginNotImplemented error = fmt.Errorf("plugin not implemented")

func (s *PluginHandlerSymbol) Version() string {
	if s.sVersion == nil {
		return ""
	}
	defer utils.CatchPanic()

	return s.sVersion.(func() string)()
}

func (s *PluginHandlerSymbol) Signature() string {
	if s.sSignature == nil {
		return ""
	}
	defer utils.CatchPanic()

	return s.sSignature.(func() string)()
}

func (s *PluginHandlerSymbol) ExportedData() *PluginParamsOut {
	if s.sExportedData == nil {
		return nil
	}
	defer utils.CatchPanic()

	return s.sExportedData.(func() *PluginParamsOut)()
}

func (s *PluginHandlerSymbol) Init(in *PluginParamsIn) error {
	if s.sInit == nil {
		return errPluginNotImplemented
	}
	defer utils.CatchPanic()

	return s.sInit.(func(*PluginParamsIn) error)(in)
}

func (s *PluginHandlerSymbol) Close() error {
	if s.sClose == nil {
		return errPluginNotImplemented
	}
	defer utils.CatchPanic()

	return s.sClose.(func() error)()
}

func (s *PluginHandlerSymbol) RequestOTP(req *common.NhpOTPRequest, helper *NhpServerPluginHelper) error {
	if s.sRequestOTP == nil {
		return errPluginNotImplemented
	}
	defer utils.CatchPanic()

	return s.sRequestOTP.(func(*common.NhpOTPRequest, *NhpServerPluginHelper) error)(req, helper)
}

func (s *PluginHandlerSymbol) RegisterAgent(req *common.NhpRegisterRequest, helper *NhpServerPluginHelper) (*common.ServerRegisterAckMsg, error) {
	if s.sRegisterAgent == nil {
		return nil, errPluginNotImplemented
	}
	defer utils.CatchPanic()

	return s.sRegisterAgent.(func(*common.NhpRegisterRequest, *NhpServerPluginHelper) (*common.ServerRegisterAckMsg, error))(req, helper)
}

func (s *PluginHandlerSymbol) ListService(req *common.NhpListRequest, helper *NhpServerPluginHelper) (*common.ServerListResultMsg, error) {
	if s.sListService == nil {
		return nil, errPluginNotImplemented
	}
	defer utils.CatchPanic()

	return s.sListService.(func(*common.NhpListRequest, *NhpServerPluginHelper) (*common.ServerListResultMsg, error))(req, helper)
}

func (s *PluginHandlerSymbol) AuthWithNHP(req *common.NhpAuthRequest, helper *NhpServerPluginHelper) (*common.ServerKnockAckMsg, error) {
	if s.sAuthWithNHP == nil {
		log.Error("AuthWithNHP not implemented")
		return nil, errPluginNotImplemented
	}
	defer utils.CatchPanic()

	return s.sAuthWithNHP.(func(*common.NhpAuthRequest, *NhpServerPluginHelper) (*common.ServerKnockAckMsg, error))(req, helper)
}

func (s *PluginHandlerSymbol) AuthWithHttp(ctx *gin.Context, req *common.HttpKnockRequest, hlpr *HttpServerPluginHelper) (*common.ServerKnockAckMsg, error) {
	if s.sAuthWithHttp == nil {
		return nil, errPluginNotImplemented
	}
	defer utils.CatchPanic()

	return s.sAuthWithHttp.(func(*gin.Context, *common.HttpKnockRequest, *HttpServerPluginHelper) (*common.ServerKnockAckMsg, error))(ctx, req, hlpr)
}

func ReadPluginHandler(pluginPath string) PluginHandler {
	p, err := plugin.Open(filepath.Join(ExeDirPath, "plugins", pluginPath))
	if err != nil {
		log.Error("open plugin %s failed: %v", pluginPath, err)
		return nil
	}
	s := &PluginHandlerSymbol{}
	s.sVersion, _ = p.Lookup("Version")
	s.sSignature, _ = p.Lookup("Signature")
	s.sExportedData, _ = p.Lookup("ExportedData")
	s.sInit, _ = p.Lookup("Init")
	s.sClose, _ = p.Lookup("Close")
	s.sRequestOTP, _ = p.Lookup("RequestOTP")
	s.sRegisterAgent, _ = p.Lookup("RegisterAgent")
	s.sListService, _ = p.Lookup("ListService")
	s.sAuthWithNHP, _ = p.Lookup("AuthWithNHP")
	s.sAuthWithHttp, _ = p.Lookup("AuthWithHttp")

	return s
}

type PluginParamsIn struct {
	PluginDirPath *string
	Log           *log.Logger
	Hostname      *string
	LocalIp       *string
	LocalMac      *string
}

type PluginParamsOut struct {
}

type NhpPluginPostAuthFunc func(*common.NhpAuthRequest, *common.ResourceData) (*common.ServerKnockAckMsg, error)

type HttpPluginPostAuthFunc func(*common.HttpKnockRequest, *common.ResourceData) (*common.ServerKnockAckMsg, error)

type NhpServerPluginHelper struct {
	StopSignal              <-chan struct{}
	AuthWithNhpCallbackFunc NhpPluginPostAuthFunc
}

type HttpServerPluginHelper struct {
	StopSignal               <-chan struct{}
	AuthWithHttpCallbackFunc HttpPluginPostAuthFunc
}
