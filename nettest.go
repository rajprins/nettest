//----------------------------------------------------------------------------
// nettest.go
// Tests network availability based on a yaml configuration file.
//
// It supports the testing of both TCP and HTTP endpoints.
// Once testing completes, a report is generated.
//----------------------------------------------------------------------------
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// Version number of this application
const version string = "1.0"

// Output report file name
const logfile string = "testresults.log"

// Template used to generate output report
const templText string = `
=========================================================================================
[TEST DETAILS]
	Network tested         : {{.Request.NetworkName}}
	Endpoint requested     : {{.Request.Proto}}://{{.Request.Host}}:{{.Request.Port}}{{.Request.Path}}
	Connected successfully : {{.Success}}
	Total request time     : {{.Time}}
	Failure Message        : {{.FailureMessage}}

[HTTP ONLY DETAILS]
	HTTP Status code       : {{.Status}}
	IP-DNS resolution      : {{.IPResolvedStatus}}
	Response body          : {{.Body}}
`

// outputDirectory is the location to write nettest output report.
var outputDirectory string

// logFlag enables the printing of the nettest output report to stdOut.
var logFlag bool

// configLocation specifies where the nettest config (yaml) can be found;
// default is `resources/config.yaml`
var configLocation string

// timeout overrides the timeout for all tcp & http requests. When unset,
// the timeout in the nettest config (yaml) is respected.
var timeout int

// versionFlag instructs nettest to print the version and exit.
var versionFlag bool

func init() {
	flag.StringVar(&configLocation, "config", "resources/config.yaml", "Location of the nettest config file. Accepts a local file location or a HTTP web server location.")
	flag.StringVar(&outputDirectory, "directory", ".", "Directory to save the nettest report.")
	flag.BoolVar(&logFlag, "log", false, "Prints test report to standard out.")
	flag.IntVar(&timeout, "timeout", 10, "Timeout for all test endpoints. If not specified, setting in nettest config file is respected. If no value was specified in the nettest config for the given endpoint, the default is used.")
	flag.BoolVar(&versionFlag, "version", false, "Print version and exit.")
	flag.Parse()
}

func main() {
	if versionFlag == true {
		fmt.Printf("Nettest version %s\n", version)
		return
	}

	intro()

	config, err := parseConfig(configLocation)
	if err != nil {
		log.Fatal("Program exiting due to nettest config file error.")
	}

	results := runTests(config)
	generateReport(config, results)
}

func intro() {
	fmt.Printf("┌────────────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ %sN E T T E S T%s                                                          │\n", CLR_RED, CLR_N)
	fmt.Printf("│ Network Connectivity Testing Utility v%s                              │\n", version)
	fmt.Printf("└────────────────────────────────────────────────────────────────────────┘\n")
}

// runTests is the entry point for test requests. It routes HTTP and
// TCP requests to their respective function. It then appends results
// to an array of ResponseDetails it has created, which is will then
// return once all tests have completed.
func runTests(config TestConfig) []ResponseDetails {
	var results []ResponseDetails

	fmt.Printf("Running test '%s'\n", config.TestName)
	for _, netTest := range config.Config {
		resp := ResponseDetails{}
		lowerCaseProto := strings.ToLower(netTest.Proto)
		if lowerCaseProto == "http" || lowerCaseProto == "https" {
			resp = testHTTPConnection(netTest)
		} else if lowerCaseProto == "tcp" {
			resp = testTCPConnection(netTest)
		} else {
			failureCause := fmt.Sprintf("Protocol \"%s\" specified for host \"%s\" is invalid. Must be tcp, http, or https.", lowerCaseProto, netTest.Host)
			log.Printf(failureCause)
			resp.Request = netTest
			resp.FailureMessage = failureCause
		}

		results = append(results, resp)
	}
	return results
}

// testHTTPConnection is responsible for testing http connection using go's http client.
func testHTTPConnection(test Configuration) ResponseDetails {
	fmt.Printf("> Host: %s (%s)... ", test.Host, strings.ToUpper(test.Proto))

	if test.Timeout == 0 {
		test.Timeout = timeout
	}

	startTime := time.Now()
	respDetails := ResponseDetails{Request: test, Success: false}

	if !strings.HasPrefix(test.Path, "/") {
		test.Path = "/" + test.Path
	}

	url := fmt.Sprintf("%s://%s:%d%s", test.Proto, test.Host, test.Port, test.Path)
	req, errRequest := http.NewRequest("GET", url, nil)
	if errRequest != nil {
		log.Printf("\n[ERROR] Unable to generate HTTP request for test %s. %s\n", test.NetworkName, errRequest.Error())
		respDetails.FailureMessage = errRequest.Error()
		return respDetails
	}

	ipRes, err := net.ResolveIPAddr("ip4", test.Host)
	respDetails.IPResolvedStatus = ipRes.String()

	if err != nil {
		respDetails.IPResolvedStatus = "Failed to resolve IP from DNS."
	}

	client := http.Client{Timeout: time.Duration(test.Timeout) * time.Second}
	resp, errClientReq := client.Do(req)

	if errClientReq != nil {
		log.Printf("\n[ERROR] Unable to access host: %s\n", errClientReq.Error())
		respDetails.FailureMessage = errClientReq.Error()
	} else {
		defer resp.Body.Close()
		respDetails.Status = resp.StatusCode
		respDetails.Success = true

		if test.CaptureBody == true {
			respBody, err := ioutil.ReadAll(resp.Body)
			respDetails.Body = string(respBody)
			if err != nil {
				log.Printf("\n[ERROR] Unable to capture response body: %s\n", err.Error())
			} else {
				fmt.Printf("[%sOK%s]\n", CLR_GREEN, CLR_N)
			}
		} else {
			fmt.Printf("[%sOK%s]\n", CLR_GREEN, CLR_N)
		}

	}

	respDetails.Time = time.Since(startTime).String()
	return respDetails
}

// testTCPConnection is responsible for testing tcp connection using tcp.Dial.
func testTCPConnection(test Configuration) ResponseDetails {
	fmt.Printf("> Host: %s (%s)... ", test.Host, strings.ToUpper(test.Proto))

	if test.Timeout == 0 {
		test.Timeout = timeout
	}

	startTime := time.Now()
	respDetails := ResponseDetails{Request: test, Success: false}
	_, err := net.DialTimeout("tcp", net.JoinHostPort(test.Host, strconv.Itoa(test.Port)), time.Duration(test.Timeout)*time.Second)

	if err != nil {
		log.Printf("\n[ERROR] Failed to access host via TCP: %s\n", err.Error())
		respDetails.FailureMessage = err.Error()
	} else {
		respDetails.Success = true
		fmt.Printf("[%sOK%s]\n", CLR_GREEN, CLR_N)
	}

	respDetails.Time = time.Since(startTime).String()
	return respDetails
}

// generateReport uses a go template to create a file detailing all the
// test results, thata were contained in the ResponseDetails slice.
func generateReport(test TestConfig, results []ResponseDetails) {
	if !strings.HasSuffix(outputDirectory, "/") {
		outputDirectory += "/"
	}

	resultOutput, err := os.Create(outputDirectory + logfile)

	if err != nil {
		log.Fatalf("Failed to create file. Exiting. %s", err.Error())
	}

	fmt.Fprintf(resultOutput, "Test name: %s", test.TestName)

	tmpl, err := template.New("outputReport").Parse(templText)
	if err != nil {
		log.Fatalf("Failed to parse output template. Exiting. %s", err.Error())
	}

	for _, testResult := range results {
		tmpl.Execute(resultOutput, testResult)
		if logFlag == true {
			tmpl.Execute(os.Stdout, testResult)
		}
	}
	fmt.Printf("\nNetwork test(s) complete.\nPlease check file %s for more details.\n\n", logfile)
}
