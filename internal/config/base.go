package config

import (
	"errors"
	"regexp"
)

type UpstreamType int

type RouteMatch struct {
	Prefix string
	CompiletedRegex *regexp.Regexp
}

type Route struct {
	Match RouteMatch
	Upstream string
}

type VirtualHost struct {
	Name string
	Domain []string
	Routes []Route
}

type Listener struct {
	Name, Address string
	Virtualhosts []VirtualHost
}

type Upstream struct {
	Name string
	Address string
	Type UpstreamType
}

type All struct {
	Listeners []Listener
	Upstream []Upstream
}

const (
	__INVALID_UPSTREAM     UpstreamType = iota
	UPSTREAM_GRPC
	UPSTREAM_HTTP
)

func (t *UpstreamType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var yamlSequence string
	unmarshal(&yamlSequence)

	switch yamlSequence {
	case "grpc":
		*t = UPSTREAM_GRPC
	case "http":
		*t = UPSTREAM_HTTP
	default:
		return errors.New("Unknown UpstreamType " + yamlSequence)
	}

	return nil
}

func (t *RouteMatch) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var yamlSequence map[string]string
	err := unmarshal(&yamlSequence)

	t.CompiletedRegex = regexp.MustCompile(yamlSequence["prefix"])
	t.Prefix = yamlSequence["prefix"]
	return err
}