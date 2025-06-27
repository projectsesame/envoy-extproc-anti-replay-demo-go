package aswatson

import (
	"regexp"
)

const (
	DefaultSsoPath    = "/admin/user/sso"
	DefaultHost       = "http://rbac-asw-authorization-blue.rbac-aswatson-prd:8080"
	DefaultHostPrefix = "http://"
	HttpsHostPrefix   = "https://"
	DefaultPort       = "80"
	HttpsPort         = "443"
	DefaultScheme     = "http"
	HttpsScheme       = "https"
)

const (
	KHost                 = "host"
	KPath                 = "path"
	KTimeout              = "timeout"
	KForgetMatchingLength = "forgetMatchingLength"
	KSoapaction           = "soapaction"
	KOriPath              = "originalPath"
	KRplPath              = "replacePath"
	KPathExtractRegexp    = "pathExtractRegexp"
	KPathSplitRegexp      = "pathSplitRegexp"
	KCheckPath            = "checkPath"
	KNoCheckRegexp        = "noCheckRegexp"
)

var ProtocolIPPortRegexp = regexp.MustCompile(`^(?:(?P<protocol>https?)://)?(?P<host>[a-zA-Z0-9._-]+)(?::(?P<port>\d+))?(?P<path>/\S*)?$`)

var http2To1Header = map[string]string{
	":authority": "host",
	":path":      "path",
	":method":    "method",
	":scheme":    "scheme",
}
