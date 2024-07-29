package aswatson

import (
	"fmt"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	ep "github.com/wrossmorrow/envoy-extproc-sdk-go"
	"io"
	"net/http"
	"os"
	"strings"
)

func NewHost(host string) string {
	if strings.HasPrefix(host, DefaultHostPrefix) {
		return host
	}

	match := ProtocolIPPortRegexp.FindStringSubmatch(host)
	if match != nil {
		regHost := match[ProtocolIPPortRegexp.SubexpIndex("host")]
		regPort := match[ProtocolIPPortRegexp.SubexpIndex("port")]
		if regPort == "" {
			regPort = DefaultPort
		}
		return fmt.Sprintf("%s%s:%s", DefaultHostPrefix, regHost, regPort)

	}
	return DefaultHost
}

func NewPath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return fmt.Sprintf("/%s", path)
}

func ReadBody(response *http.Response) []byte {
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	return body
}

func Cancel(code int32, errMsg string, ctx *ep.RequestContext) error {
	if errMsg == "" {
		errMsg = typev3.StatusCode_name[code]
	}
	ctx.CancelRequest(code, map[string]ep.HeaderValue{}, errMsg)
	return fmt.Errorf(errMsg)
}

func GetEnv(key string, defaultValue string) string {

	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return value
}

func ToHeaders(m map[string][]byte) http.Header {
	headers := http.Header{}
	for k, v := range m {
		if http2To1Header[k] != "" {
			headers.Set(http2To1Header[k], string(v))
		} else {
			headers.Set(k, string(v))
		}
	}
	return headers
}
