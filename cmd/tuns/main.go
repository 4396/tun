package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/4396/tun/log"
	"github.com/4396/tun/server"
	"github.com/4396/tun/version"
	"gopkg.in/ini.v1"
)

type tunServer struct {
	*server.Server
	proxies *ini.File
}

func newServer(conf string) (s *tunServer, err error) {
	cfg, err := ini.Load(conf)
	if err != nil {
		return
	}

	sec := cfg.Section("tuns")
	addr := sec.Key("addr").MustString(":7000")
	http := sec.Key("http").MustString(":7070")

	s = &tunServer{proxies: cfg}
	svr, err := server.Listen(&server.Config{
		Addr:     addr,
		AddrHTTP: http,
		Auth:     s.Auth,
		Load:     s.Load,
		TraffIn:  s.TraffIn,
		TraffOut: s.TraffOut,
	})
	if err != nil {
		return
	}

	s.Server = svr
	return
}

func (s *tunServer) Auth(id, token string) (err error) {
	sec, err := s.proxies.GetSection(id)
	if err != nil {
		err = fmt.Errorf("proxy '%s' not exists", id)
		return
	}

	if token != sec.Key("token").String() {
		err = errors.New("token does not match")
	}
	return
}

func (s *tunServer) Load(loader server.Loader, id string) (err error) {
	sec, err := s.proxies.GetSection(id)
	if err != nil {
		err = fmt.Errorf("proxy '%s' not exists", id)
		return
	}

	switch sec.Key("type").String() {
	case "tcp":
		var port int
		port, err = sec.Key("port").Int()
		if err != nil {
			return
		}
		err = loader.ProxyTCP(id, port)
	case "http":
		domain := sec.Key("domain").String()
		err = loader.ProxyHTTP(id, domain)
	default:
		err = errors.New("unexpected proxy type")
	}
	return
}

func (s *tunServer) TraffIn(id string, b []byte) {
	log.Infof("[IN] %s %d", id, len(b))
}

func (s *tunServer) TraffOut(id string, b []byte) {
	log.Infof("[OUT] %s %d", id, len(b))
}

var (
	conf = flag.String("c", "conf/tuns.ini", "config file's path")
)

func main() {
	flag.Parse()
	log.Infof("Start tun server, version is %s", version.Version)

	s, err := newServer(*conf)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(s.Run(context.Background()))
}
