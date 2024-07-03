// WEB Application Server
//
// Author: Joel Midstjärna
// License: MIT
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// Below globals are automatically updated by the CI by providing
// -X linker flags while building
var applicationVersion = "<NOT SET>"
var applicationBuildTime = "<NOT SET>"
var applicationGitHash = "<NOT SET>"

func printUsage() {
	fmt.Printf("Usage: %s [options] [<apppath>] [<datapath>]\n\n", os.Args[0])
	fmt.Printf("<apppath> is the directory where your web files are located.\n")
	fmt.Printf("Default is current directory.\n\n")
	fmt.Printf("<datapath> is the directory where data (JSON) is stored.\n")
	fmt.Printf("Default is same as apppath.\n\n")
	fmt.Printf("Supported options:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage
	var version = flag.Bool("v", false, "Display version")
	var port = flag.Int("p", 8080, "Network port to listen to")
	var tlsEnable = flag.Bool("s", false, "Use secure connection (TLS/HTTPS)")
	var tlsCertFile = flag.String("c", "cert.pem", "TLS certificate file")
	var tlsKeyFile = flag.String("k", "key.pem", "TLS key file")
	flag.Parse()

	if *version {
		fmt.Printf("Version:    %s\n", applicationVersion)
		fmt.Printf("Build Time: %s\n", applicationBuildTime)
		fmt.Printf("GIT Hash:   %s\n", applicationGitHash)
		os.Exit(0)
	}

	appDir := "."
	if flag.NArg() >= 1 {
		appDir = flag.Arg(0)
	} 
	dataDir := appDir
	if flag.NArg() >= 2 {
		dataDir = flag.Arg(1)
	} 
	if flag.NArg() > 2 {
		fmt.Fprintf(os.Stderr, "Invalid number of arguments!\n\n")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("apppath %s, datapath %s\n", appDir, dataDir)

	fs := http.FileServer(http.Dir(appDir))
	http.Handle("/", fs)

	if *tlsEnable {
		fmt.Printf("Serving path %s on port %d over HTTPS\n", appDir, *port)

		err := http.ListenAndServeTLS(":"+strconv.Itoa(*port), *tlsCertFile, *tlsKeyFile, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error! %s\n", err)
		}
	} else {
		fmt.Printf("Serving path %s on port %d over HTTP\n", appDir, *port)

		err := http.ListenAndServe(":"+strconv.Itoa(*port), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error! %s\n", err)
		}	
	} 
}
