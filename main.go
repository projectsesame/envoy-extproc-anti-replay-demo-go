package main

import (
	"flag"
	"github.com/projectsesame/envoy-extproc-anti-replay-demo-go/aswatson/adapter"
	"github.com/projectsesame/envoy-extproc-anti-replay-demo-go/aswatson/path"
	"github.com/projectsesame/envoy-extproc-anti-replay-demo-go/aswatson/sso"
	"github.com/projectsesame/envoy-extproc-anti-replay-demo-go/aswatson/token"
	"log"
	"os"

	ep "github.com/wrossmorrow/envoy-extproc-sdk-go"
)

type processor interface {
	Init(args []string, set *flag.FlagSet, opts *ep.ProcessingOptions) error
	Finish()
	ep.RequestProcessor
}

var processors = map[string]processor{
	"watson-auth-sso":     &sso.WatsonSSORequestProcessor{},
	"watson-auth-sso-2":   &sso.WatsonSSORequestProcessorV2{},
	"watson-auth-path":    &path.WatsonPathRequestProcessor{},
	"watson-auth-adapter": &adapter.WatsonAdapterRequestProcessor{},
	"watson-auth-token":   &token.WatsonTokenRequestProcessor{},
}

func parseArgs(args []string, port *int, opts *ep.ProcessingOptions) *flag.FlagSet {
	rootCmd := flag.NewFlagSet("root", flag.ExitOnError)
	rootCmd.IntVar(port, "port", 50051, "the gRPC port.")

	//opts = ep.NewDefaultOptions()

	rootCmd.BoolVar(&opts.LogStream, "log-stream", false, "log the stream or not.")
	rootCmd.BoolVar(&opts.LogPhases, "log-phases", false, "log the phases or not.")
	rootCmd.BoolVar(&opts.UpdateExtProcHeader, "update-extproc-header", false, "update the extProc header or not.")
	rootCmd.BoolVar(&opts.UpdateDurationHeader, "update-duration-header", false, "update the duration header or not.")

	return rootCmd
}

func main() {
	// cmd subCmd arg, arg2,...
	args := os.Args
	if len(args) < 2 {
		log.Fatal("Passing a processor is required.")
	}

	cmd := args[1]
	proc, exists := processors[cmd]
	if !exists {
		log.Fatalf("Processor \"%s\" not defined.", cmd)
	}
	var port int
	opts := ep.NewDefaultOptions()
	rootCmd := parseArgs(os.Args[2:], &port, opts)
	var err error
	if err = proc.Init(os.Args[2:], rootCmd, opts); err != nil {
		log.Fatalf("Initialize the processor is failed: %v.", err.Error())
	}
	defer proc.Finish()

	ep.Serve(port, proc)
}
