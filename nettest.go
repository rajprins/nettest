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
[GENERIC TEST RESULTS]
Network tested         : {{.Request.NetworkName}}
Endpoint requested     : {{.Request.Proto}}://{{.Request.Host}}:{{.Request.Port}}{{.Request.Path}}
Connected successfully : {{.Success}}
Total request time     : {{.Time}}
Failure Message        : {{.FailureMessage}}

[HTTP ONLY RESULTS]
HTTP Status code       : {{.Status}}
IP-DNS resolution      : {{.IPResolvedStatus}}
Response body          : {{.Body}}
`

// outputDirectory is the location to write nettest output report.
var outputDirectory string

// logFlag enables the printing of the nettest output report to stdOut.
var logFlag bool

// configLocation specifies where the nettest config (yaml) can be found;
// default is `config.yaml`
var configLocation string

// timeout overrides the timeout for all tcp & http requests. When unset,
// the timeout in the nettest config (yaml) is respected.
var timeout int

// versionFlag instructs nettest to print the version and exit.
var versionFlag bool

func init() {
	flag.StringVar(&configLocation, "config", "config.yaml", "Location of the nettest config file. Accepts a local file location or a HTTP web server location.")
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
		fmt.Printf("\nProgram exiting due to nettest config file error.\nPlease check your configuration file for errors.\n")
		os.Exit(1)
	}

	results := runTests(config)
	generateReport(config, results)
}

func intro() {
	fmt.Printf("┌────────────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ %s// N E T T E S T //%s                                                    │\n", CLR_RED, CLR_N)
	fmt.Printf("│ Network Connectivity Testing Utility v%s                              │\n", version)
	fmt.Printf("└────────────────────────────────────────────────────────────────────────┘\n")
}

// runTests is the entry point for test requests. It routes HTTP and
// TCP requests to their respective function. It then appends results
// to an array of ResponseDetails it has created, which is will then
// return once all tests have completed.
func runTests(config TestConfig) []ResponseDetails {
	var results []ResponseDetails

	fmt.Printf("\n%sRunning test suite: %s%s\n", CLR_WHITE, config.TestName, CLR_N)
	fmt.Printf("──────────────────────────────────────────────────────────────────────────\n")

	for testNr, netTest := range config.Config {
		resp := ResponseDetails{}
		lowerCaseProto := strings.ToLower(netTest.Proto)
		if lowerCaseProto == "http" || lowerCaseProto == "https" {
			resp = testHTTPConnection(testNr, netTest)
		} else if lowerCaseProto == "tcp" {
			resp = testTCPConnection(testNr, netTest)
		} else {
			failureCause := fmt.Sprintf("[%d] Configurstion error: protocol \"%s\" specified for host \"%s\" is invalid. Must be TCP, HTTP, or HTTPS.\n", testNr, lowerCaseProto, netTest.Host)
			fmt.Printf(failureCause)
			resp.Request = netTest
			resp.FailureMessage = failureCause
		}

		results = append(results, resp)
	}
	return results
}

// testHTTPConnection is responsible for testing http connection using go's http client.
func testHTTPConnection(testNr int, test Configuration) ResponseDetails {
	fmt.Printf("[%d] Target: %s (%s)... ", (testNr + 1), test.Host, strings.ToUpper(test.Proto))

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
		fmt.Printf("[%sFAILED%s]\nUnable to generate HTTP request for test %s. %s\n", CLR_RED, CLR_N, test.NetworkName, errRequest.Error())
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
		fmt.Printf("[%sFAILED%s]\nUnable to access target: %s\n", CLR_RED, CLR_N, errClientReq.Error())
		respDetails.FailureMessage = errClientReq.Error()
	} else {
		defer resp.Body.Close()
		respDetails.Status = resp.StatusCode
		respDetails.Success = true

		if test.CaptureBody == true {
			respBody, err := ioutil.ReadAll(resp.Body)
			respDetails.Body = string(respBody)
			if err != nil {
				fmt.Printf("[%sFAILED%s]\nUnable to capture response body: %s\n", CLR_RED, CLR_N, err.Error())
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
func testTCPConnection(testNr int, test Configuration) ResponseDetails {
	fmt.Printf("[%d] Target: %s (%s)... ", (testNr + 1), test.Host, strings.ToUpper(test.Proto))

	if test.Timeout == 0 {
		test.Timeout = timeout
	}

	startTime := time.Now()
	respDetails := ResponseDetails{Request: test, Success: false}
	_, err := net.DialTimeout("tcp", net.JoinHostPort(test.Host, strconv.Itoa(test.Port)), time.Duration(test.Timeout)*time.Second)

	if err != nil {
		fmt.Printf("[%sFAILED%s]\nUnable to access host via TCP: %s\n", CLR_RED, CLR_N, err.Error())
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
		fmt.Printf("Error. Failed to create file %s: %s", logfile, err.Error())
		os.Exit(1)
	}

	fmt.Fprintf(resultOutput, "Test suite: %s", test.TestName)

	tmpl, err := template.New("outputReport").Parse(templText)
	if err != nil {
		fmt.Printf("Failed to parse output template. Exiting. %s", err.Error())
		os.Exit(1)
	}

	for _, testResult := range results {
		tmpl.Execute(resultOutput, testResult)
		if logFlag == true {
			tmpl.Execute(os.Stdout, testResult)
		}
	}
	fmt.Printf("\nNetwork test(s) complete.\nPlease check file %s for more details.\n\n", logfile)
}
