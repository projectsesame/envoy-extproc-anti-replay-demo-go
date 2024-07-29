package aswatson

import (
	"regexp"
)

const (
	DefaultSsoPath    = "/admin/user/sso"
	DefaultHost       = "rbac-asw-authorization-blue.rbac-aswatson-prd:8080"
	DefaultHostPrefix = "http://"
	DefaultPort       = "80"
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

var ProtocolIPPortRegexp = regexp.MustCompile(`^(?P<protocol>(https|http)?://)?(?P<host>([a-zA-Z0-9_-]+)*(\.[a-zA-Z0-9_-]+)*)(:(?P<port>\d+))?((?P<path>/[^\s]*)*)?$`)

var http2To1Header = map[string]string{
	":authority": "host",
	":path":      "path",
	":method":    "method",
	":scheme":    "scheme",
}
