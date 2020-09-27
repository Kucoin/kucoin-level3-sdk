package api

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strings"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/app"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/cfg"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/log"
	"go.uber.org/zap"
)

//Server is api server
type Server struct {
	app *app.App
}

func fixMarketName(market string) string {
	return strings.Split(market, "_v")[0]
}

//InitRpcServer init rpc server
func InitRpcServer(app *app.App) {
	apiAddress := cfg.AppConfig.ApiServer.Address

	if err := rpc.Register(&Server{
		app: app,
	}); err != nil {
		log.Panic("rpc Register error: " + err.Error())
	}

	log.Info("start running rpc server, listen: " + cfg.AppConfig.ApiServer.Network + "://" + apiAddress)

	if strings.HasPrefix(cfg.AppConfig.ApiServer.Network, "unix") {
		err := os.Remove(apiAddress)
		switch {
		case os.IsNotExist(err), err == nil:
		default:
			log.Panic("remove socket failed", zap.Error(err))
		}
	}
	listener, err := net.Listen(cfg.AppConfig.ApiServer.Network, apiAddress)
	if err != nil {
		log.Panic("api server run failed, error: " + err.Error())
	}
	defer func() { _ = listener.Close() }()
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go jsonrpc.ServeConn(conn)
	}
}

//TokenMessage is token type message
type TokenMessage struct {
	Token string `json:"token"`
}

//Response is api response
type Response struct {
	Code  string      `json:"code"`
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func (s *Server) checkToken(token string) *Response {
	if token != cfg.AppConfig.ApiServer.Token {
		resp := s.failure(TokenErrorCode, "error rpc token")
		return &resp
	}

	return nil
}

func (s *Server) success(data interface{}) Response {
	return Response{
		Code:  "0",
		Data:  data,
		Error: "",
	}
}

const (
	ServerErrorCode = "10"
	TokenErrorCode  = "20"
	TickerErrorCode = "30"
	ConfNotFound    = "40"
)

func (s *Server) failure(code string, err string) Response {
	return Response{
		Code:  code,
		Data:  "",
		Error: err,
	}
}
