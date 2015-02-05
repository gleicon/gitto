// Copyright 2014 gitto authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func (s *httpServer) route() {

	http.HandleFunc("/api/v1/github/hook", s.githubHookHandler)
	http.HandleFunc("/api/v1/status", s.statusHandler)
	http.HandleFunc("/api/v1/requests/", s.requestsHandler)

	// Static file server.
	http.Handle("/", http.FileServer(http.Dir(s.config.DocumentRoot)))
}

func (s *httpServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, world\r\n")
}

func (s *httpServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	appc := make(map[string]map[string]string)
	for _, application := range s.config.Application {
		an := application.Name
		cc, err := s.metrics.GetCounters(an)
		if err != nil {
			log.Println(err)
		}
		appc[an] = cc
	}
	jsonStatus, err := json.Marshal(appc)
	if err != nil {
		httpError(w, r, http.StatusInternalServerError, err)
		return
	}
	fmt.Fprintf(w, string(jsonStatus))
}

func (s *httpServer) requestsHandler(w http.ResponseWriter, r *http.Request) {
	appname := r.URL.Path[len("/api/v1/requests/"):]
	if len(appname) < 1 {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	rr, err := s.metrics.GetRequests(appname, 10)

	if err != nil {
		httpError(w, r, http.StatusInternalServerError, err)
		return
	}

	jsonStatus, err := json.Marshal(rr)
	if err != nil {
		httpError(w, r, http.StatusInternalServerError, err)
		return
	}
	fmt.Fprintf(w, string(jsonStatus))
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

		go handlePush(ref, repository, name, s.config.Application, s.metrics)
		fmt.Fprintf(w, "OK\r\n")
		return
	}
	http.Error(w, "Method not allowed", 405)
	return
}
