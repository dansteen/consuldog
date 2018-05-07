# consuldog
Zero-conf, consul based, service discovery daemon for DataDog inspired by fabio ([This fabio](http://github.com/fabiolb/fabio) not [this fabio](http://www.fabioinc.com/)).


## The Problem
[Datadog](http://datadoghq.com) is a monitoring solution for servers and services, however, it relies on static files to determin *what* it needs to monitor.  This is fine if you are running a couple services on a box in a fairly static way.  However, if you are running services inside a cluster manager you never know what services are on what box are listening on what IP and on what port.  Monitoring the right things at the right time becomes more difficult. 

## The solution
Enter consuldog.  Consuldog is a daemon that listens for service changes in [consul](https://www.consul.io/), generates datadog check.d/*.yaml files based on templates, and tells datadog to rescan its config files.   Consuldog is inspired by [fabio](http://github.com/fabiolb/fabio) and attempts to be zero-conf in the same fashion.   


# Usage
## Invocation
consuldog does not require any command line flags to run sucessfully (assuming the defaults work for you):
```
./consuldog
```
That's it!

## Details
Consuldog relies on specific tags being present in consul on the services you wish to monitor.   Each service you want to monitor should have a tag set in the following format:
```
<prefix>:<template_uri>:<datadog_config_name>
```
Where:

| word                  | Usage                                                                                                                                                                                                                     | Default         |
|-----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------|
| \<prefix>             | A freeform string that lets consuldog know that this is a service that needs to be monitored.                                                                                                                             | consuldogConfig |
| <template_uri>       | The uri to the template consuldog should *ingest* to generate datadog configs for this services. |  n/a             |
| <datadog_config_name> | The name of the datadog config file to generate for this service, using the template mentioned above.  This should just be the base name of the config without the `.yaml` extension (e.g. `apache`, *not* `apache.yaml`) | n/a             |

More, concretely, if you are using the default prefix, the tag would look like this:
```
consuldogConfig:http://myhost.com/app_apache.yaml:apache
```
consuldog would recognize the above tag as indicating that this is a service datadog should monitor, and attempt to get `http://myhost.com/app_apache.yaml`.  It would then use that template as part of the datadog config file named apache.yaml (note that you don't specify the filename extension for the datadog_config_name).  If there are multiple services that generate the same datadog config (e.g. multiple apache services) all of them would be merged into a single apache.yaml file for datadog to use.

## Templates
Templates are golang templates, and the general structure of the templates *must* match the standard config files that datadog provides.  Specifically, the templates are expected to have the format:
```
init_config:
  - value
  - value
instances:
  - item: 1
    item: 2
```
Where init_config contains config items you would like to see in the final datadog yaml file, and instances are monitoring instances for your service.  
In the template, the following template variables will be replaced with their respective values from the consul service:

| Template Variable  | Definition                                                        | Format   |
|--------------------|-------------------------------------------------------------------|----------|
| {{ .Address }}     | The IP address the service is listening on                        | string   |
| {{ .Port }}        | The Port the services is listening on                             | int      |
| {{ .Service }}     | The Name of the service                                           | string   |
| {{ .ID }}          | A globally unique ID for the service (this is usually quite long) | string   |
| {{ .Tags }}        | A list of tags on this service                                    | []string |
| {{ .CreateIndex }} | The CreateIndex of the service (this is a consul thing)           | uint64   |
| {{ .ModifyIndex }} | The ModifyIndex of the service (this is a consul thing)           | uint64   |

### Examples
The following will generate monitoring for apache:
```
init_config:
instances:
  - apache_status_url: http://{{.Address}}:{{.Port}}/server-status?auto
```
and, assuming two apache services on the box will generate the following config:
```
init_config:
instances:
  - apache_status_url: http://10.10.20.56:45322/server-status?auto
  - apache_status_url: http://10.10.20.56:76232/server-status?auto
```
a config instance is generated for each instance of the service on that particular box.


## Command line switches
consuldog can run without any configuration, and will monitor services that are correctly tagged, and that have templates.  However, should you wish, there are a number of tunable items:

| Short Flag | Long Flag                  | Can be passed multiple times | Function                                                                                                                                                                                                                                               |
|------------|----------------------------|------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| -a         | --consulAddress            | no                           | the address of the consul agent (default "http://localhost:8500")                                                                                                                                                                                      |
| -d         | --datadogFolder            | no                           | the base datadog config folder (the one containing the datadog.conf file) (default "/etc/dd-agent")                                                                                                                                                    |
| -m         | --datadogMinReloadInterval | no                           | the minimum number of seconds between reloads of the DataDog process regardless of how many times the configs are updated in that time. (default 10)                                                                                                   |
| -k         | --datadogProcName          | no                           | the name of the datadog process we should send reload signals to.,A process with this name that is running as the same user as consuldog (if one can be found) will be sent a HUP signal when new datadog configs are written. (default "supervisord") |
| -n         | --nodeName                 | yes                          | the name of the node we want to look at the services of (default is the name of the node of the consul agent we are connecting to)                                                                                                                     |
| -p         | --prefix                   | no                           | the consul tag prefix to look for in consul to know that a service needs monitoring (default "consuldogConfig")                                                                                                                                       |
| -t         | --tempFolder           | no                           | the folder to user for temporary file storage |
