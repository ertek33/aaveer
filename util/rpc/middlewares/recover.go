package middlewares

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/openweb3/go-rpc-provider"
	"github.com/sirupsen/logrus"
)

var (
	errMiddlewareCrashed = errors.New("RPC middleware crashed")
)

func Recover(next rpc.HandleCallMsgFunc) rpc.HandleCallMsgFunc {
	return func(ctx context.Context, msg *rpc.JsonRpcMessage) (resp *rpc.JsonRpcMessage) {
		defer func() {
			if err := recover(); err != nil {
				resp = msg.ErrorResponse(errMiddlewareCrashed)

				debug.PrintStack()

				logrus.WithFields(logrus.Fields{
					"inputMsg": newHumanReadableRpcMessage(msg),
					"panicErr": err,
				}).Error("RPC middleware panic recovered")
			}
		}()

		return next(ctx, msg)
	}
}

type humanReadableRpcMessage struct {
	Version string
	ID      string
	Method  string
	Params  string
	Error   error
	Result  string
}

func newHumanReadableRpcMessage(msg *rpc.JsonRpcMessage) *humanReadableRpcMessage {
	return &humanReadableRpcMessage{
		ID:      string(msg.ID),
		Version: msg.Version,
		Method:  msg.Method,
		Params:  string(msg.Params),
		Error:   msg.Error,
		Result:  string(msg.Result),
	}
}
