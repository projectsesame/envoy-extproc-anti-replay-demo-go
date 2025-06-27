package aswatson

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	ep "github.com/wrossmorrow/envoy-extproc-sdk-go"
)

func NewHost(host string) string {
	if strings.HasPrefix(host, DefaultHostPrefix) || strings.HasPrefix(host, HttpsHostPrefix) {
		return host
	}

	match := ProtocolIPPortRegexp.FindStringSubmatch(host)
	if match != nil {
		group := func(name string) string {
			idx := ProtocolIPPortRegexp.SubexpIndex(name)
			if idx >= 0 && idx < len(match) {
				return match[idx]
			}
			return ""
		}
		regHost := group("host")
		regPort := group("port")
		regProtocol := group("protocol")
		if regPort == "" && regProtocol == "" {
			regProtocol = DefaultScheme
			regPort = DefaultPort
		} else if regPort == "" {
			switch regProtocol {
			case HttpsScheme:
				regPort = HttpsPort
			default:
				regPort = DefaultPort
			}
		} else if regProtocol == "" {
			switch regPort {
			case HttpsPort:
				regProtocol = HttpsScheme
			default:
				regProtocol = DefaultScheme
			}
		}
		return fmt.Sprintf("%s://%s:%s", regProtocol, regHost, regPort)
	}
	return DefaultHost
}

func HttpClient(timeout time.Duration, reqUrl string) *http.Client {
	u, _ := url.Parse(reqUrl)
	if u.Scheme == HttpsScheme {
		return &http.Client{
			Timeout:   timeout,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}
	}
	return &http.Client{
		Timeout: timeout,
	}
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
