package io

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/grafana/grafanactl/internal/agent"
	"github.com/grafana/grafanactl/internal/format"
	"github.com/spf13/pflag"
)

type Options struct {
	OutputFormat string

	customCodecs  map[string]format.Codec
	defaultFormat string
}

func (opts *Options) RegisterCustomCodec(name string, codec format.Codec) {
	if opts.customCodecs == nil {
		opts.customCodecs = make(map[string]format.Codec)
	}

	opts.customCodecs[name] = codec
}

func (opts *Options) DefaultFormat(name string) {
	opts.defaultFormat = name
}

func (opts *Options) BindFlags(flags *pflag.FlagSet) {
	defaultFormat := "json"
	if opts.defaultFormat != "" {
		defaultFormat = opts.defaultFormat
	}

	// Agent mode: override any per-command default with JSON.
	// Explicit -o flag from user still takes precedence (via cobra flag parsing).
	if agent.IsAgentMode() {
		defaultFormat = "json"
	}

	flags.StringVarP(&opts.OutputFormat, "output", "o", defaultFormat, "Output format. One of: "+strings.Join(opts.allowedCodecs(), ", "))
}

func (opts *Options) Validate() error {
	codec := opts.codecFor(opts.OutputFormat)
	if codec == nil {
		return fmt.Errorf("unknown output format '%s'. Valid formats are: %s", opts.OutputFormat, strings.Join(opts.allowedCodecs(), ", "))
	}

	return nil
}

// Codec returns the codec for the configured output format.
// We have to return an interface here.
func (opts *Options) Codec() (format.Codec, error) { //nolint:ireturn
	codec := opts.codecFor(opts.OutputFormat)
	if codec == nil {
		return nil, fmt.Errorf(
			"unknown output format '%s'. Valid formats are: %s", opts.OutputFormat, strings.Join(opts.allowedCodecs(), ", "),
		)
	}

	return codec, nil
}

func (opts *Options) Encode(dst io.Writer, value any) error {
	codec, err := opts.Codec()
	if err != nil {
		return err
	}

	return codec.Encode(dst, value)
}

// We have to return an interface here.
func (opts *Options) codecFor(format string) format.Codec { //nolint:ireturn
	if opts.customCodecs != nil && opts.customCodecs[format] != nil {
		return opts.customCodecs[format]
	}

	return opts.builtinCodecs()[format]
}

func (opts *Options) builtinCodecs() map[string]format.Codec {
	return map[string]format.Codec{
		"yaml": format.NewYAMLCodec(),
		"json": format.NewJSONCodec(),
	}
}

func (opts *Options) allowedCodecs() []string {
	allowedCodecs := slices.Collect(maps.Keys(opts.builtinCodecs()))
	for name := range opts.customCodecs {
		allowedCodecs = append(allowedCodecs, name)
	}

	// the allowed codecs are stored in a map: let's sort them to make the
	// return value of this function deterministic
	sort.Strings(allowedCodecs)

	return allowedCodecs
}
