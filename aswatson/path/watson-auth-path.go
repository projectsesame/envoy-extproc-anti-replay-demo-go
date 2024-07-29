package path

import (
	"flag"
	"github.com/projectsesame/envoy-extproc-anti-replay-demo-go/aswatson"
	ep "github.com/wrossmorrow/envoy-extproc-sdk-go"
	"log"
	"net/http"
	"regexp"
)

const (
	defaultMatchReg     = `^/mdm/supplier/(.*)$`
	defaultPathSplitReg = `[^/]+`
	defaultCheckPath    = "costs"
)

var PathMatchRegexp *regexp.Regexp
var PathSplitRegexp *regexp.Regexp

type WatsonPathRequestProcessor struct {
	opts *ep.ProcessingOptions

	ExtractRegexp string
	SplitRegexp   string
	CheckPath     string
}

func (w *WatsonPathRequestProcessor) GetName() string {
	return "waston-auth-path"
}

func (w *WatsonPathRequestProcessor) GetOptions() *ep.ProcessingOptions {
	return w.opts
}

func (w *WatsonPathRequestProcessor) ProcessRequestHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	if s := PathMatchRegexp.FindString(ctx.Path); s != "" {
		for i := range PathSplitRegexp.FindAllString(s, -1) {
			if PathSplitRegexp.FindAllString(s, -1)[i] == w.CheckPath {
				log.Printf("request [%s] path check is failed", ctx.RequestID)
				aswatson.Cancel(http.StatusForbidden, "Security policy authentication failed", ctx)
				return nil
			}
		}
	}
	log.Printf("request [%s] path check is success", ctx.RequestID)
	return ctx.ContinueRequest()
}

func (w *WatsonPathRequestProcessor) ProcessRequestBody(ctx *ep.RequestContext, body []byte) error {
	return ctx.ContinueRequest()

}

func (w *WatsonPathRequestProcessor) ProcessRequestTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonPathRequestProcessor) ProcessResponseHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonPathRequestProcessor) ProcessResponseBody(ctx *ep.RequestContext, body []byte) error {
	return ctx.ContinueRequest()
}

func (w *WatsonPathRequestProcessor) ProcessResponseTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (w *WatsonPathRequestProcessor) Init(args []string, rootCmd *flag.FlagSet, opts *ep.ProcessingOptions) error {

	w.opts = opts

	rootCmd.StringVar(&w.ExtractRegexp, aswatson.KPathExtractRegexp, aswatson.GetEnv("PATH_EXTRACT_REGEXP", defaultMatchReg), "the path match regexp.")
	rootCmd.StringVar(&w.SplitRegexp, aswatson.KPathSplitRegexp, aswatson.GetEnv("PATH_SPLIT_REGEXP", defaultPathSplitReg), "the path split regexp.")
	rootCmd.StringVar(&w.CheckPath, aswatson.KCheckPath, aswatson.GetEnv("CHECK_PATH", defaultCheckPath), "the check path.")
	rootCmd.Parse(args)

	log.Printf("path extract regexp: %s", w.ExtractRegexp)
	PathMatchRegexp = regexp.MustCompile(w.ExtractRegexp)
	log.Printf("path split regexp: %s", w.SplitRegexp)
	PathSplitRegexp = regexp.MustCompile(w.SplitRegexp)
	log.Printf("check path: %s", w.CheckPath)
	return nil
}

func (w *WatsonPathRequestProcessor) Finish() {

}
