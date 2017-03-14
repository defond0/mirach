/*
Package compinfo is a plugin that provides information about the asset.

Calling via the CLI

To run this plugin from the command line interface:

	mirach compinfo
	# or
	mirach compinfo --infogroup load

to get an individual group of information.

For full usage information run:

	mirach compinfo --help

Calling via the API

To use this plugin via the API:

	info.<InfoGroup>.GetInfo()
	// returns: info.<InfoGroup>
	// or
	info.<InfoGroup>.String()
	// returns a string
	// ex. fmt.Println(info.<InfoGroup>)

There are several other ways to call via the API, but this is the most simple
and will include all information collected.
*/
package compinfo
