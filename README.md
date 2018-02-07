# nettest

Config-based network connectivity tester. Setup a YAML file describing targets you'd like to test connections to. This tool was created from the need to test a customer's network before arriving. For example, needing to test connectivity to various locations before delivering trainings or workshops.

## Quick start

* Download the nettest compressed archive for [MacOS](https://github.com/rajprins/nettest/releases/download/v1.0/nettest-macos.tar.gz), 
[Linux](https://github.com/rajprins/nettest/releases/download/v1.0/nettest-linux.tar.gz), 
or [Windows](https://github.com/rajprins/nettest/releases/download/v1.0/nettest-windows.zip).

* Unpack the archive, which contains the nettest binary and a pre-configured config file.

  ```
  $ tar zxvf nettest-${PLATFORM}.tar.gz
  $ cd ${PLATFORM}
  $ ls -l
  
  nettest
  config.yaml 
  ```
  > Replace ${PLATFORM} with either "linux" or "macos". If you are using Windows, unzip the archive in .zip format.

* If desired, modify the config.yaml. See [Configuration file](#configuration-file) for more details.

* Run the test.

	```
	./nettest
	```

  > You can also point to a nettest config (yaml) hosted on an http server. See the `-config` flag for more details.

* Examine the test report.

	```
	cat testresults.log
	```

	```
	Test name: Network Test Valid 1
	=========================================================================================
	[GENERIC TEST RESULTS]
	Network tested         : google
	Endpoint requested     : http://google.com:80
	Connected successfully : true
	Total request time     : 88.212457ms
	Failure Message        : 

	[HTTP ONLY RESULTS]
	HTTP Status code       : 200
	IP-DNS resolution      : 172.217.19.206
	Response body          : 

	=========================================================================================
	[GENERIC TEST RESULTS]
	Network tested         : amazon
	Endpoint requested     : https://amazon.com:443
	Connected successfully : true
	Total request time     : 1.388307603s
	Failure Message        : 

	[HTTP ONLY RESULTS]
	HTTP Status code       : 200
	IP-DNS resolution      : 176.32.98.166
	Response body          : 
	```

* See [Usage](#usage) and [Configuration file](#configuration-file) for more details..

## Compiling from source

This section assumes you have a [proper Go environment configured](https://golang.org/doc/install).

1. Retrieve nettest and navigate to its directory.

	```
	$ go get github.com/rajprins/nettest

	$ cd $GOPATH/src/github.com/rajprins/nettest
	```

1. Compile `nettest` for your [target architecture](https://golang.org/doc/install/source#environment).
	```bash
	# example, compiling for windows (.exe) from non-windows box	
	GOOS=windows GOARCH=386 go build -o nettest.exe *.go
	```
	> If your target arch is the same as the machine you're compiling on, simply omit `GOOS` and `GOARCH` above.

1. Create a config file based on [the sample config](./config.yaml).

1. Upload the `config.yaml` and `nettest` executable to the same directory in the desired testing machine.

## Usage

`nettest` requires a [config file](#configuration-file) to understand how to run. By default, `nettest` will look for a file named `config.yaml` in your current working directory and write the test report in the same directory as a file named `testresults.log`. To change these behaviors and see other usage options, run `nettest -h`.

```
Usage of ./nettest:
  -config string
    	Location of the nettest config file. Accepts a local file location or a HTTP web server location. (default "config.yaml")
  -directory string
    	Directory to save the nettest report. (default ".")
  -log
    	Prints test report to standard out.
  -timeout int
    	Timeout for all test endpoints. If not specified, setting in nettest config file is respected. If no value was specified in the nettest config for the given endpoint, the default is used. (default 10)
  -version
    	Print version and exit.

```

## Configuration file

Config is achieved in YAML. An example config is as follows.

```yaml
testname: Network Test Valid 1
email: john.doe@email.com
config:

  - networkname: google
    host: google.com
    port: 80
    path:
    proto: http
    capturebody: false

  - networkname: amazon
    host: amazon.com
    port: 443
    path:
    proto: https
    capturebody: false
```

Each key and its purpose is as follows.

- `testname`: Name of the overall network test case.
- `email`: Email address the report should be sent to. This tool doesn't send the email, but it provides a means of letting the operator (one running the code) see where the output should be sent.
- `config`: List of networks you'd like to test.
	- `networkname`: Name of the network you're testing.
	- `host`: Host name or IP to reach out to.
	- `port`: Port name associated with host.
	- `path`: Path to append to host when making a request. Always start with a `/`. Include desired query parameter(s).
	- `proto`: Protocol, `tcp`, `http` or `https`.
	- `timeout`: Time in seconds a request should wait before failing.
	- `captureBody`: Whether you'd like the response for the endpoint included in the final report.

## Contributing

Feel free to open issues / PRs.
