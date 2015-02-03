// Copyright 2014 gitto authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func (s *httpServer) route() {

	http.HandleFunc("/api/v1/github/hook", s.githubHookHandler)
	http.HandleFunc("/api/v1/status", s.testHandler)
	http.HandleFunc("/api/v1/pause", s.testHandler)
	http.HandleFunc("/api/v1/apps", s.testHandler)

	// Static file server.
	http.Handle("/", http.FileServer(http.Dir(s.config.DocumentRoot)))
}

func (s *httpServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, world\r\n")
}

func (s *httpServer) githubHookHandler(w http.ResponseWriter, r *http.Request) {
	ipPort := strings.Split(r.RemoteAddr, ":")
	if checkIPAddr(ipPort[0]) == false {
		// invalid origin
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		fmt.Fprintf(w, "OK\r\n")
		return
	}

	if r.Method == "POST" {
		// process webhook
		event := r.Header.Get("X-GitHub-Event")
		if event == "" {
			http.Error(w, "No events", http.StatusNotFound)
			return
		}
		if event == "ping" {
			fmt.Fprintf(w, `{"msg" : "Hi!"}`)
			return
		}
		if event != "push" {
			fmt.Fprintf(w, `{"msg" : "Bad event type"}`)
			return

		}

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			httpError(w, r, http.StatusInternalServerError, err)
			return
		}

		data := make(map[string]interface{})

		err = json.Unmarshal([]byte(body), &data)
		if err != nil {
			httpError(w, r, http.StatusInternalServerError, err)
			return
		}

		r := data["repository"].(map[string]interface{})
		n := r["owner"].(map[string]interface{})

		ref := data["ref"].(string)
		repository := r["name"].(string)
		name := n["name"].(string)

		go handlePush(ref, repository, name, s.config.Application)
		fmt.Fprintf(w, "OK\r\n")
		return
	}
	http.Error(w, "Method not allowed", 405)
	return
}

func (s *httpServer) testHandler(w http.ResponseWriter, r *http.Request) {
	var v string
	var err error
	if v, err = s.redis.Get("hello"); err != nil {
		httpError(w, r, 503, err)
		return
	}

	fmt.Fprintf(w, "hello %s\r\n", v)
}
