package token

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/projectsesame/envoy-extproc-anti-replay-demo-go/aswatson"
	"log"
	"net/http"
	"time"

	ep "github.com/wrossmorrow/envoy-extproc-sdk-go"
)

const (
	defaultTokenPath = "/oauth/check_token"
	defaultTokenHost = "rbac-asw-authorization-blue.rbac-aswatson-prd:8080"
)

type WatsonTokenRequestProcessor struct {
	opts *ep.ProcessingOptions

	Host       string
	Path       string
	Timeout    time.Duration
	HTTPClient *http.Client
	TokenURL   string
}

func (w *WatsonTokenRequestProcessor) tokenURL() string {
	return fmt.Sprintf("%s%s", aswatson.NewHost(w.Host), aswatson.NewPath(w.Path))
}

func (w *WatsonTokenRequestProcessor) tokenAuth(ctx context.Context, token string, rawHeaders map[string][]byte) (*http.Response, error) {
	body := fmt.Sprintf("token=%s", token)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, w.TokenURL, bytes.NewBuffer([]byte(body)))
	req.Header = aswatson.ToHeaders(rawHeaders)

	req.Host = w.Host
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	return w.HTTPClient.Do(req)
}

func (w *WatsonTokenRequestProcessor) GetName() string {
	return "watson-auth-token"
}

func (w *WatsonTokenRequestProcessor) GetOptions() *ep.ProcessingOptions {
	return w.opts
}

func (w *WatsonTokenRequestProcessor) ProcessRequestHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	if err := w.tokenCheck(ctx); err != nil {
		log.Printf("request [%s] token check is failed", ctx.RequestID)
		return nil
	}
	return ctx.ContinueRequest()
}

func (w *WatsonTokenRequestProcessor) tokenCheck(ctx *ep.RequestContext) error {
	token := ctx.AllHeaders.RawHeaders["token"]
	if token == nil {
		return aswatson.Cancel(http.StatusUnauthorized, "token is not found", ctx)
	}

	resp, err := w.tokenAuth(context.Background(), string(token), ctx.AllHeaders.RawHeaders)
	if err != nil {
		return aswatson.Cancel(http.StatusInternalServerError, err.Error(), ctx)
	}

	if resp.StatusCode == http.StatusOK {
		headerMap := map[string]ep.HeaderValue{
			"content-type": {RawValue: []byte("application/json;charset=UTF-8")},
			":authority":   {RawValue: []byte("tv-chart.aswatson.net")},
		}
		ctx.SetValue("header", headerMap)
		ctx.SetValue("body", tokenBody(resp))
		//ctx.ReplaceBodyChunk(tokenBody(resp))
		return nil
		//return ctx.OverwriteHeaders(headerMap)
	}
	return aswatson.Cancel(int32(resp.StatusCode), string(aswatson.ReadBody(resp)), ctx)
}

func tokenBody(resp *http.Response) []byte {
	body := aswatson.ReadBody(resp)
	tempMap := map[string]interface{}{}
	json.Unmarshal(body, &tempMap)
	userName := tempMap["user_name"]

	return []byte(fmt.Sprintf(`{"username": "%s"}`, userName))

}

func (w *WatsonTokenRequestProcessor) ProcessRequestBody(ctx *ep.RequestContext, body []byte) error {
	headersMap, err := ctx.GetValue("header")
	if err != nil {
		log.Printf(aswatson.Cancel(http.StatusUnauthorized, err.Error(), ctx).Error())
		return nil
	}
	b, err := ctx.GetValue("body")
	if err != nil {
		log.Printf(aswatson.Cancel(http.StatusUnauthorized, err.Error(), ctx).Error())
		return nil
	}
	ctx.ReplaceBodyChunk(b.([]byte))
	ctx.OverwriteHeaders(headersMap.(map[string]ep.HeaderValue))

	log.Printf("request [%s] token check body set is success", ctx.RequestID)
	return ctx.ContinueRequest()
}

func (w *WatsonTokenRequestProcessor) ProcessRequestTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonTokenRequestProcessor) ProcessResponseHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonTokenRequestProcessor) ProcessResponseBody(ctx *ep.RequestContext, body []byte) error {
	return ctx.ContinueRequest()
}

func (w *WatsonTokenRequestProcessor) ProcessResponseTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonTokenRequestProcessor) Init(args []string, rootCmd *flag.FlagSet, opts *ep.ProcessingOptions) error {

	w.opts = opts

	rootCmd.StringVar(&w.Host, aswatson.KHost, aswatson.GetEnv("TOKEN_HOST", defaultTokenHost), "the host of the token service.")
	rootCmd.StringVar(&w.Path, aswatson.KPath, aswatson.GetEnv("TOKEN_PATH", defaultTokenPath), "the path of the token service.")
	rootCmd.DurationVar(&w.Timeout, aswatson.KTimeout, 10*time.Second, "the timeout of the http client.")
	rootCmd.Parse(args)

	w.TokenURL = w.tokenURL()
	w.HTTPClient = &http.Client{Timeout: w.Timeout}

	log.Printf("tokenURL: %s", w.TokenURL)
	log.Printf("timeout: %s", w.Timeout)
	return nil
}

func (w *WatsonTokenRequestProcessor) Finish() {

}
