package sso

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/projectsesame/envoy-extproc-anti-replay-demo-go/aswatson"
	"log"
	"net/http"
	"regexp"
	"time"

	ep "github.com/wrossmorrow/envoy-extproc-sdk-go"
)

const defaultNoCheckReg = `.*swagger%-ui.html$`

var NoCheckRegexp *regexp.Regexp

type WatsonSSORequestProcessorV2 struct {
	opts *ep.ProcessingOptions

	Host                 string
	Path                 string
	ForgetMatchingLength int
	Timeout              time.Duration
	HTTPClient           *http.Client
	SsoURL               string

	NoCheckReg string
}

func (w *WatsonSSORequestProcessorV2) newRequestBody(rawHeaders map[string][]byte) *RequestBody {

	return &RequestBody{
		RequestURL:           string(rawHeaders[":path"]),
		RequestMethod:        string(rawHeaders[":method"]),
		ForgetMatchingLength: w.ForgetMatchingLength,
	}
}

func (w *WatsonSSORequestProcessorV2) ssoURL() string {
	return fmt.Sprintf("%s%s", aswatson.NewHost(w.Host), aswatson.NewPath(w.Path))
}

func (w *WatsonSSORequestProcessorV2) ssoAuth(ctx context.Context, rawHeaders map[string][]byte) (*http.Response, error) {
	body := w.newRequestBody(rawHeaders)
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, w.SsoURL, bytes.NewBuffer(jsonBody))
	req.Header = aswatson.ToHeaders(rawHeaders)
	req.Host = w.Host
	req.Header.Set("content-type", "application/json")

	return w.HTTPClient.Do(req)
}

func (w *WatsonSSORequestProcessorV2) ssoCheck(ctx *ep.RequestContext, rawHeaders map[string][]byte) error {
	resp, err := w.ssoAuth(context.Background(), rawHeaders)
	if err != nil {
		return aswatson.Cancel(http.StatusInternalServerError, err.Error(), ctx)
	}

	if resp.StatusCode == http.StatusOK {
		headerMap := map[string]ep.HeaderValue{
			"userid":   {RawValue: []byte(resp.Header.Get("userid"))},
			"username": {RawValue: []byte(resp.Header.Get("username"))},
		}

		return ctx.AddHeaders(headerMap)
	}

	return aswatson.Cancel(int32(resp.StatusCode), string(aswatson.ReadBody(resp)), ctx)

}

func (w *WatsonSSORequestProcessorV2) GetName() string {
	return "watson-auth-sso"
}

func (w *WatsonSSORequestProcessorV2) GetOptions() *ep.ProcessingOptions {
	return w.opts
}

func (w *WatsonSSORequestProcessorV2) ProcessRequestHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	fmt.Printf(NoCheckRegexp.FindString(string(headers.RawHeaders[":path"])))
	if NoCheckRegexp.FindString(string(headers.RawHeaders[":path"])) == "" && NoCheckRegexp.FindString(string(headers.RawHeaders["referer"])) == "" {
		if err := w.ssoCheck(ctx, headers.RawHeaders); err != nil {
			log.Printf("request [%s] sso check is failed", ctx.RequestID)
			return nil
		}
		log.Printf("request [%s] sso check is success", ctx.RequestID)
		return ctx.ContinueRequest()
	}
	log.Printf("request [%s] skip sso check", ctx.RequestID)
	return ctx.ContinueRequest()

}

func (w *WatsonSSORequestProcessorV2) ProcessRequestBody(ctx *ep.RequestContext, body []byte) error {
	return ctx.ContinueRequest()
}

func (w *WatsonSSORequestProcessorV2) ProcessRequestTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonSSORequestProcessorV2) ProcessResponseHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonSSORequestProcessorV2) ProcessResponseBody(ctx *ep.RequestContext, body []byte) error {
	return ctx.ContinueRequest()
}

func (w *WatsonSSORequestProcessorV2) ProcessResponseTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonSSORequestProcessorV2) Init(args []string, rootCmd *flag.FlagSet, opts *ep.ProcessingOptions) error {

	w.opts = opts

	rootCmd.StringVar(&w.Host, aswatson.KHost, aswatson.GetEnv("SSO_HOST", defaultHost), "the host of the sso service.")
	rootCmd.StringVar(&w.Path, aswatson.KPath, aswatson.GetEnv("SSO_PATH", defaultSsoPath), "the path of the sso service.")
	rootCmd.IntVar(&w.ForgetMatchingLength, aswatson.KForgetMatchingLength, defaultForgetMatchingLength, "the forget matching length.")
	rootCmd.DurationVar(&w.Timeout, aswatson.KTimeout, 10*time.Second, "the timeout of the http client.")
	rootCmd.StringVar(&w.NoCheckReg, aswatson.KNoCheckRegexp, aswatson.GetEnv("NO_CHECK_REG", defaultNoCheckReg), "the no check regexp.")
	rootCmd.Parse(args)

	w.SsoURL = w.ssoURL()
	w.HTTPClient = &http.Client{Timeout: w.Timeout}

	NoCheckRegexp = regexp.MustCompile(w.NoCheckReg)

	log.Printf("ssourl: %s", w.SsoURL)
	log.Printf("forgetMatchingLength: %d", w.ForgetMatchingLength)
	log.Printf("timeout: %s", w.Timeout)
	log.Printf("noCheckReg: %s", w.NoCheckReg)
	return nil
}

func (w *WatsonSSORequestProcessorV2) Finish() {

}
