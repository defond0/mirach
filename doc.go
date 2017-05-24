/*
Package mirach is a tool to get information about a machine and send it to a central repository.

This is a tool to be installed on a machine about which one needs to gather data.
Broadly, it opens a connection with an MQTT broker, AWS IoT by default, and
using that connection, sends data collected about the machine to a central
clearing house. It's aim is to be very simple to implement and very secure.
To that end, it uses two client certificates to authenticate with the broker.
One, the customer certificate, is used to create a client that can only be used
to register an asset. The other, the asset certificate, is used for nearly all
remaining publications and subscriptions.

Features

	- command line interface
	- built in data collection
		- package information (installed / available)
		- computer information
	- support for custom data collection plugins
	- overrides for builtin plugins
	- plugin load can be delayed to prevent overloading
	- communication over MQTT
	- client authenticated TLS

Installation

The easiest way to install a go package from source is using the go get command.

	go get github.com/cleardataeng/mirach

But if you don't need the source you can download the binary for your operating
system and architecture from _insert links here_.

Usage

Running mirach is simple:

	mirach

To see usage instructions, run:

	mirach --help

Or for detailed information on subcommands:

	mirach [command] --help
	# for example
	mirach compinfo --help

Configuration

For mirach to work, it needs to authenticate with the MQTT broker. It does so
using client certificates. You need a client certificate and private key on your
machine for mirach to register the asset. You configuration can be stored in any
of three places: the same directory as the executable binary, the user directory
of the user running mirach, or the system-wide configuration directory. These
are searched in that order.

On Windows:

	1. current directory
	2. c:\Users\<user>\AppData\Local\mirach\
	3. c:\ProgramData\mirach\

On Linux:

	1. current directory
	2. ~/.config/mirach/
	3. /etc/mirach/

The file structure in any (or a combination) of these folders is:

	- .../mirach/
		- keys/
			- asset/
				- keys/
					- ca.pem.crt (automatic)
					- private.pem.key (automatic)
			- customer/
				- keys/
					- ca.pem.crt (get from ClearDATA)
					- private.pem.key (get from ClearDATA)
		- config.(hcl,json,prop,yaml) (automatic, but can be customized)

Here is a sample configuration in yaml:

	asset:
	  id: 12345678-asset-1
	customer:
	  id: "12345678"
	plugins:
	  custom:
	    custom_plugin_1:
	      cmd: custom command
	      schedule: '@every 5m'
	      load_delay: '30s'
	    custom_plugin_2:
	      cmd: other_cmd
	      schedule: "0 30 * * * *"
	  builtin:
	    compinfo-load:
	      schedule: '@every 15s'
	      load_delay: '30s'

Notes

mirach will need to run as a user that has permissions to list installed and
available packages, discover computer information like host and CPU data, make
TCP connections, and any further permissions required by your custom plugins.

When your asset client certificates are revoked or lost, mirach will attempt to
re-register the asset. If, at that time, the customer client certificate is
still valid, and new asset certificate will be issue, downloaded, and used. If
your client certificate is revoked or lost a replacement needs to be requested.

The decision when overriding values given precedence to the value given in the
configuration file. For boolean values that default to false, the truth table
look like the following.

code  | override  | result
------|-----------|-----------
false | None      | false (default)
true  | None      | true
false | true      | true
true  | true      | true
false | false     | false
true  | false     | false

Credits

	- Jeffrey DeFond jeff.defond@cleardata.com
	- Herkermer Sherwood herkermer.sherwood@cleardata.com
*/
package main
