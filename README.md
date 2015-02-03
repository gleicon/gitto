## gitto ?

Smartass webhook handler for github repos. 
This application listens to github webhook events and execute pre-configured actions.

## configuration format

  The configuration format is based on TOML (https://github.com/toml-lang/toml)

		debug = false
		templates_dir = "./assets/templates"
		document_root = "./assets/public_html"

		[http_server]
		addr = ":8080"
		xheaders = false

		[https_server]
		#addr = ":8081"
		#cert_file = "./ssl/server.crt"
		#key_file = "./ssl/server.key"

		[[application]]
		name = "imager"
		repo = "gleicon/imager"
		path = "/tmp/"
		init_command = "git clone"
		sync_command = "git pull"
		post_command = "/bin/true"

		[[application]]
		name = "imager2"
		repo = "gleicon/imager"
		path = "/tmp/app2/"
		init_command = "git clone"
		sync_command = "git pull"
		post_command = "/bin/true"

	For each project you want to monitor you need to create a new [[aplication]] session.
	Beyond name and repo, which will be prepended by github.com, you need to configure a destination path and 3 commands.
	If a directory with "name" doesn't exists inside "path" gitto will run "init_command" for the first time.
	For each push event sync_command will be executed inside "path"+/"+"repo" and subsequently "post_command" will be executed.


## private repositories
	- follow github's public key procedure
	- create the key for the user you will be running gitto
	- you can test manually trying to clone the repo 

## Preparing the environment to build

Prerequisites:

- Git
- rsync
- GNU Make
- [Go](http://golang.org) 1.0.3 or newer

Install dependencies, and compile:

	make deps
	make clean
	make all

Generate a self-signed SSL certificate (optional):

	cd ssl
	make

Edit the config file and run the server:

	vi gitto.conf
	./gitto

Install, uninstall. Edit Makefile and set PREFIX to the target directory:

	sudo make install
	sudo make uninstall

Allow non-root process to listen on low ports:

	/sbin/setcap 'cap_net_bind_service=+ep' /opt/gitto/server

Good luck!
