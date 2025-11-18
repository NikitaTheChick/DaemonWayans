package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/namsral/flag"
)

const defaultTick = 60 * time.Second

type config struct {
	contentType string
	server      string
	statusCode  int
	tick        time.Duration
	url         string
	userAgent   string
}

/*
Takes command line arguments as input, builds FlagSet(set of defined flags), lists and parses each flag,
assigns flag values to config
*/
func (c *config) init(args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.String(flag.DefaultConfigFlagname, "", "Path to config file")

	var contentType = flags.String("content_type", "", "Content-Type HTTP header value")
	var server = flags.String("server", "", "Server HTTP header value")
	var statusCode = flags.Int("status", 200, "Response HTTP status code")
	var tick = flags.Duration("tick", defaultTick, "Ticking interval")
	var url = flags.String("url", "", "Request URL")
	var userAgent = flags.String("user_agent", "", "User-Agent HTTP header value")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	c.contentType = *contentType
	c.server = *server
	c.statusCode = *statusCode
	c.tick = *tick
	c.url = *url
	c.userAgent = *userAgent

	return nil

}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	c := &config{}

	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	go func() {
		for {
			select {
			case <-signalChan:
				log.Printf("Got SIGINT/SIGTERM. Later fam!")
				cancel()
				os.Exit(1)
			case <-ctx.Done():
				log.Printf("Dunzo.")
				os.Exit(1)
			}
		}
	}()

	if err := run(ctx, c, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, c *config, out io.Writer) error {
	c.init(os.Args)
	log.SetOutput(out)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(c.tick):
			resp, err := http.Get(c.url)
			if err != nil {
				return err
			}

			if resp.StatusCode != c.statusCode {
				log.Printf("Them status codes don't match, fam. Got: %d\n", resp.StatusCode)
			}

			if s := resp.Header.Get("server"); s != c.server {
				log.Printf("Them server headers don't match, fam. Got: %s\n", s)
			}

			if ct := resp.Header.Get("content-type"); ct != c.contentType {
				log.Printf("Them content-type headers don't match, fam. Got: %s\n", ct)
			}

			if ua := resp.Header.Get("user-agent"); ua != c.userAgent {
				log.Printf("Them user-agent headers don't match, fam. Got: %s\n", ua)
			}
		}
	}
}
