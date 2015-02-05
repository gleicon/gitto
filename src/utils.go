// Copyright 2014 gitto authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

func execCmd(cmd string, repo string) ([]byte, error) {
	cmds := strings.Split(cmd, " ")
	cmds = append(cmds, repo)
	log.Println(cmds)
	c := cmds[0]
	out, err := exec.Command(c, cmds[1:]...).Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func handlePush(ref string, repository string, name string, applications []application) {
	log.Printf("Ref: %s Repository: %s Name:%s\n", repository, ref, name)
	repo := fmt.Sprintf("%s/%s", name, repository)

	for _, app := range applications {
		if app.Repo == repo {
			destDir := fmt.Sprintf("%s/%s", app.Path, repository)
			src, err := os.Stat(destDir)
			if err != nil {
				log.Println(err)
			}
			gh_repo := fmt.Sprintf("https://github.com/%s", repo)

			if src != nil {
				if !src.IsDir() {
					log.Println("Not a directory")
					return
				}
				log.Printf("Chdir to %s\n", destDir)
				os.Chdir(destDir)
				out, err := execCmd(app.SyncCommand, gh_repo)
				if err != nil {
					log.Println(err)
				}
				log.Printf("cmd output %s", strings.TrimFunc(string(out), unicode.IsSpace))
			} else {
				log.Printf("Chdir to %s", app.Path)
				os.Chdir(app.Path)
				out, err := execCmd(app.InitCommand, gh_repo)
				if err != nil {
					log.Println(err)
				}
				log.Println(string(out))
			}
			out, err := execCmd(app.PostCommand, destDir)
			if err != nil {
				log.Println(err)
			}
			log.Println(string(out))

		}
	}
}

// check if origin ip is contained at some of gh ranges
func checkIPAddr(IPAddr string) bool {
	ip := net.ParseIP(IPAddr)
	if ip == nil {
		return false
	}
	for _, ipn := range GHIP {
		r := ipn.Contains(ip)

		if r {
			return r
		}
	}
	return false
}

// goroutine to check origin IPs from github
func getGHIPs(GHIP *[]net.IPNet) {
	for {
		resp, err := http.Get("https://api.github.com/meta")
		if err != nil {
			log.Fatalln(err)
		}

		defer resp.Body.Close()

		var data struct{ Hooks []string }
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			log.Fatalln(err)
		}

		hooks := data.Hooks
		hooks = append(hooks, "127.0.0.0/8")

		log.Printf("github origins: %s\n", hooks)

		var ipn []net.IPNet

		for _, s := range hooks {
			_, n, err := net.ParseCIDR(s)
			if err != nil {
				log.Fatalln(err)
			}
			ipn = append(ipn, *n)
		}

		*GHIP = ipn
		time.Sleep(300 * time.Second)
	}
}

// remoteIP returns the remote IP without the port number.
func remoteIP(r *http.Request) string {
	// If xheaders is enabled, RemoteAddr might be a copy of
	// the X-Real-IP or X-Forwarded-For HTTP headers, which
	// can be a comma separated list of IPs. In this case,
	// only the first IP in the list is used.
	if strings.Index(r.RemoteAddr, ",") > 0 {
		r.RemoteAddr = strings.SplitN(r.RemoteAddr, ",", 2)[0]
	}
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
		return ip
	} else {
		return r.RemoteAddr
	}
}

// serverURL returns the base URL of the server based on the current request.
func serverURL(config *configFile, r *http.Request, preferSSL bool) string {
	var (
		addr  string
		host  string
		port  string
		proto string
	)
	if config.HTTPS.Addr == "" || !preferSSL {
		proto = "http"
		addr = config.HTTP.Addr
	} else {
		proto = "https"
		addr = config.HTTPS.Addr
	}
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			port = addr[i+1:]
			break
		}
	}
	host = r.Host
	if port != "" {
		for i := len(host) - 1; i >= 0; i-- {
			if host[i] == ':' {
				host = host[:i]
				break
			}
		}
		if port != "80" && port != "443" {
			host += ":" + port
		}
	}
	return fmt.Sprintf("%s://%s", proto, host)
}
