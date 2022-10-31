package rpc

import (
	"context"
	"io/ioutil"
	"time"
	"os"
	"net/http"
	"net/url"
	"encoding/json"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/patrickmn/go-cache"

	web3Types "github.com/openweb3/web3go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/scroll-tech/rpc-gateway/util/rpc/handlers"
)

var (
	whiteListURL     string
	whiteListCache = cache.New(10 * time.Minute, 10 * time.Minute)
)

func init() {
	whiteListURL = os.Getenv("WHITELIST_BACKEND_URL")
	logrus.Info("whiteListURL: ", whiteListURL)
}

// debugAPI provides ethereum debug API.
type debugAPI struct{}

func (api *debugAPI) TraceTransaction(ctx context.Context, blockHash common.Hash, opts interface{}) (*web3Types.TraceTransactionResult, error) {
	remoteAddr := remoteAddrFromContext(ctx)
	logrus.Info("remoteAddr: ", remoteAddr)
	valid, err := isIPValid(remoteAddr)
	if err != nil {
		return nil, err
	}
	if !valid {
		logrus.Debug("Invalid IP ", remoteAddr)
		return nil, errors.New("Operation not permitted")
	}
	w3c := GetEthClientFromContext(ctx)
	return w3c.Eth.TraceTransaction(blockHash, opts)
}

func remoteAddrFromContext(ctx context.Context) string {
	if ip, ok := handlers.GetIPAddressFromContext(ctx); ok {
		return ip
	}

	return "unknown_ip"
}

func isIPValid(ip string) (bool, error) {
	// in dev environment
	if whiteListURL == "" {
		return true, nil
	}

	if ip == "" {
		return false, nil
	}

	cache_key := "whitelist-ip-" + ip
	valid, found := whiteListCache.Get(cache_key)
	logrus.Debug("whitelist IP cache Get ip: ", ip, ", found: ", found, ", valid: ", valid)
	if found {
		return valid.(bool), nil
	}

    debuggerListURL, err := url.JoinPath(whiteListURL, "api/get_debugger")
    if err != nil {
        return false, err
    }

    params := url.Values{}
    Url, err := url.Parse(debuggerListURL)
    if err != nil {
        return false, err
    }

    params.Set("accept", "application/json")
	params.Set("ip", ip)
	Url.RawQuery = params.Encode()
	urlPath := Url.String()
    resp, err := http.Get(urlPath)
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return false, err
    }

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
    if err != nil {
        return false, err
    }

	debuggerList, ok := data["debugger"]
	if !ok {
		return false, errors.New("Parse Json fail, no debugger")
	}
	ok = debuggerList != nil
	whiteListCache.Set(cache_key, ok, cache.DefaultExpiration)
	return ok, nil
}
