// Copyright 2014 gitto authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	html "html/template"
	text "text/template"

	"github.com/fiorix/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

var (
	VERSION = "tip"
	APPNAME = "gitto"

	// Templates
	HTML *html.Template
	TEXT *text.Template

	// GH IPs
	GHIP []net.IPNet
)

func main() {
	configFile := flag.String("c", "gitto.conf", "")
	logFile := flag.String("l", "", "")
	flag.Usage = func() {
		fmt.Println("Usage: gitto [-c gitto.conf] [-l logfile]")
		os.Exit(1)
	}
	flag.Parse()

	var err error
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize log.
	if *logFile != "" {
		setLog(*logFile)
	}

	// Parse templates.
	HTML = html.Must(html.ParseGlob(config.TemplatesDir + "/*.html"))
	TEXT = text.Must(text.ParseGlob(config.TemplatesDir + "/*.txt"))

	// Set up databases.
	rc := redis.New(config.DB.Redis)

	// Set GOMAXPROCS and show server info.
	var cpuinfo string
	if n := runtime.NumCPU(); n > 1 {
		runtime.GOMAXPROCS(n)
		cpuinfo = fmt.Sprintf("%d CPUs", n)
	} else {
		cpuinfo = "1 CPU"
	}
	log.Printf("%s %s (%s)", APPNAME, VERSION, cpuinfo)

	// Start gh IP collection
	go getGHIPs(&GHIP)

	// Start HTTP server.
	s := new(httpServer)
	mm := NewMetrics(rc)

	s.init(config, rc, mm)
	go s.ListenAndServe()
	go s.ListenAndServeTLS()

	// Sleep forever.
	select {}
}

func setLog(filename string) {
	f := openLog(filename)
	log.SetOutput(f)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP)
	go func() {
		// Recycle log file on SIGHUP.
		var fb *os.File
		for {
			<-sigc
			fb = f
			f = openLog(filename)
			log.SetOutput(f)
			fb.Close()
		}
	}()
}

func openLog(filename string) *os.File {
	f, err := os.OpenFile(
		filename,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Fatal(err)
	}
	return f
}
