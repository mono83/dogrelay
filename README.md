dogrelay
========

Simple UDP metrics forwarder for InfluxDB with tags support
 
Compilation
-----------

Run `make release` or `go build -o dogrelay main.go`

Incoming metrics format
-----------------------

`<name>:<value>|<type>|<ignored>|#<tagName1>:<tagValue1>,...`

Example:

`cpu:94|g|@1.0|#instance:app.local,environment:live`

Metric types:

* `c` - incremental counter
* `g` - gauge
* `ms` - time duration


Command line arguments
----------------------

CLI interface uses `spf13/cobra`, so you can always use `--help` flag to obtain some help
