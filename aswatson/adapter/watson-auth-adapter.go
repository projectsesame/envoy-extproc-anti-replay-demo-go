package adapter

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/projectsesame/envoy-extproc-anti-replay-demo-go/aswatson"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	ep "github.com/wrossmorrow/envoy-extproc-sdk-go"
)

const (
	defaultGatewayHost = "dmp-gateway-cdc.cdc-aswatson-prd:10000"
	defaultOriPath     = "eai_anon_enu"
	defaultRplPath     = "member-upsert"
	defaultSoapaction  = "document/http://siebel.com/CustomUI:ASWMemberInboundMemberUpsert"
)

var SoapactionRegexp *regexp.Regexp

type WatsonAdapterRequestProcessor struct {
	opts *ep.ProcessingOptions

	Soapaction string

	Host string
	//原始路径
	OriPath string
	//替换路径
	RplPath    string
	Timeout    time.Duration
	HTTPClient *http.Client
}

func (w *WatsonAdapterRequestProcessor) GetName() string {
	return "waston-auth-adapter"
}

func (w *WatsonAdapterRequestProcessor) GetOptions() *ep.ProcessingOptions {
	return w.opts
}

func (w *WatsonAdapterRequestProcessor) ProcessRequestHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	if !w.matchPath(ctx.AllHeaders.RawHeaders) {
		log.Printf("request [%s] not match soapaction,will continue", ctx.RequestID)
		return ctx.ContinueRequest()
	}
	ctx.SetValue("soapaction", true)
	return ctx.ContinueRequest()
}

func (w *WatsonAdapterRequestProcessor) ProcessRequestBody(ctx *ep.RequestContext, body []byte) error {
	if ok, err := ctx.GetValue("soapaction"); ok.(bool) && err == nil {
		log.Printf("request [%s] match soapaction,will forward", ctx.RequestID)
		return w.forward(ctx, body)
	}
	log.Printf("request [%s] not match soapaction,will continue", ctx.RequestID)

	return ctx.ContinueRequest()
}

func (w *WatsonAdapterRequestProcessor) forwardURL(path string) string {
	return fmt.Sprintf("%s%s", aswatson.NewHost(w.Host), w.forwardPath(path))
}

func (w *WatsonAdapterRequestProcessor) forwardPath(path string) string {
	newPath := strings.ReplaceAll(path, w.OriPath, w.RplPath)
	return aswatson.NewPath(newPath)
}

func (w *WatsonAdapterRequestProcessor) forward(ctx *ep.RequestContext, body []byte) error {
	request, _ := http.NewRequest(ctx.Method, w.forwardURL(ctx.FullPath), bytes.NewBuffer(body))
	request.Header = aswatson.ToHeaders(ctx.AllHeaders.RawHeaders)
	request.Host = w.Host
	request.Header.Set(":path", strings.ReplaceAll(ctx.FullPath, w.OriPath, w.RplPath))
	do, err := w.HTTPClient.Do(request)
	if err != nil {
		return aswatson.Cancel(http.StatusInternalServerError, err.Error(), ctx)
	}
	return aswatson.Cancel(int32(do.StatusCode), string(aswatson.ReadBody(do)), ctx)
}

func (w *WatsonAdapterRequestProcessor) matchPath(headers map[string][]byte) bool {
	soapaction := string(headers["soapaction"])
	return SoapactionRegexp.MatchString(soapaction)
}

func (w *WatsonAdapterRequestProcessor) ProcessRequestTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonAdapterRequestProcessor) ProcessResponseHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonAdapterRequestProcessor) ProcessResponseBody(ctx *ep.RequestContext, body []byte) error {
	return ctx.ContinueRequest()
}

func (w *WatsonAdapterRequestProcessor) ProcessResponseTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonAdapterRequestProcessor) Init(args []string, rootCmd *flag.FlagSet, opts *ep.ProcessingOptions) error {

	w.opts = opts

	rootCmd.StringVar(&w.Host, aswatson.KHost, aswatson.GetEnv("GATEWAY_HOST", defaultGatewayHost), "the host of the gateway service.")
	rootCmd.StringVar(&w.OriPath, aswatson.KOriPath, aswatson.GetEnv("ORIGINAL_PATH", defaultOriPath), "the original path")
	rootCmd.StringVar(&w.RplPath, aswatson.KRplPath, aswatson.GetEnv("REPLACE_PATH", defaultRplPath), "the replace path")
	rootCmd.StringVar(&w.Soapaction, aswatson.KSoapaction, aswatson.GetEnv("SOAPACTION", defaultSoapaction), "the soap action")
	rootCmd.DurationVar(&w.Timeout, aswatson.KTimeout, 20*time.Second, "the timeout of the http client.")
	rootCmd.Parse(args)

	w.HTTPClient = &http.Client{Timeout: w.Timeout}

	log.Printf("forwardHost: %s", w.Host)
	log.Printf("originalPath: %s", w.OriPath)
	log.Printf("repleacePath: %s", w.RplPath)
	log.Printf("timeout: %s", w.Timeout)
	log.Printf("soapaction: %s", w.Soapaction)
	SoapactionRegexp = regexp.MustCompile(w.Soapaction)
	return nil
}

func (w *WatsonAdapterRequestProcessor) Finish() {

}
