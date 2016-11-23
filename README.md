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

Run binary file without argument to view help 

* `bind` - address to listen incoming UDP packets
* `size` - incoming packet size limit, default `4096`
* `influx` - InfluxDB UDP listening port to forward data
* `percentiles` - Percentiles to calculate, default `95,98`
* `compat` - StatsD compatibility mode. Will append `.timer`, `.counter` and `.gauge` to outgoing metrics