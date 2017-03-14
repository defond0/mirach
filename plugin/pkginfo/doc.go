/*
Package pkginfo is a plugin that provides information about installed and
available packages or KB articles (in the case of Windows).

It will also provide information about if an update is security related where
this information is provided.

Calling via the CLI

To run this plugin from the command line interface:

	mirach pkginfo

For full usage information run:

	mirach pkginfo --help

Calling via the API

To use this plugin via the API:

	pkginfo.String()

There are several other ways to call via the API, but this is the most simple
and will include all information collected.
*/
package pkginfo
