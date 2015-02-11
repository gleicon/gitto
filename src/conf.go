// Copyright 2014 gitto authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type application struct {
	Name        string `toml:"name"`
	Repo        string `toml:"repo"`
	Path        string `toml:"path"`
	InitCommand string `toml:"init_command"`
	SyncCommand string `toml:"sync_command"`
	PostCommand string `toml:"post_command"`
}

type configFile struct {
	Debug        bool   `toml:"debug"`
	TemplatesDir string `toml:"templates_dir"`
	DocumentRoot string `toml:"document_root"`
	TravisToken  string `toml:"travis_token"`

	DB struct {
		Redis string `toml:"redis"`
	} `toml:"db"`

	HTTP struct {
		Addr     string `toml:"addr"`
		XHeaders bool   `toml:"xheaders"`
	} `toml:"http_server"`

	HTTPS struct {
		Addr     string `toml:"addr"`
		CertFile string `toml:"cert_file"`
		KeyFile  string `toml:"key_file"`
	} `toml:"https_server"`

	Application []application `toml:"application"`
}

// LoadConfig reads and parses the configuration file.
func loadConfig(filename string) (*configFile, error) {
	c := &configFile{}
	if _, err := toml.DecodeFile(filename, c); err != nil {
		return nil, err
	}

	// Make files' path relative to the config file's directory.
	basedir := filepath.Dir(filename)
	relativePath(basedir, &c.DocumentRoot)
	relativePath(basedir, &c.TemplatesDir)
	relativePath(basedir, &c.HTTPS.CertFile)
	relativePath(basedir, &c.HTTPS.KeyFile)

	return c, nil
}

func relativePath(basedir string, path *string) {
	p := *path
	if p != "" && p[0] != '/' {
		*path = filepath.Join(basedir, p)
	}
}
