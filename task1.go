package main

import (
	//"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"
	"golang.org/x/net/context"
	"net/http"
	"strings"
	//"bufio"
	//"io"
	"io/ioutil"
)

type kodingRequest struct { // inherit request from http request struct
	http.Request
	name string
	operation map[string]string
}

func main() {

	Operation_ex := map[string]map[string]string{
		"check_etc_hosts_has_8888": map[string]string{
			"path": "/etc/hosts",//assume there will be a file name here too like kite.conf or tasks.txt
			"type": "file_contains",
			"check": "8.8.8.8",
		},
		"check_kite_config_file_exists": map[string]string{
			"path": "/etc/host/koding/kite.conf",
			"type": "file_exists",
		},
		"check_etc_host_runs_pid_123": map[string]string{
			"path": "/etc/hosts",
			"type": "process_runs",
			"pid": "123",
		},
	}

	for operation_name := range Operation_ex{

		hReq, err := http.NewRequest(Operation_ex[operation_name]["type"], Operation_ex[operation_name]["path"], nil) //ignore body io for now

		if(!err){
			go handler(kodingRequest{hReq, operation_name, Operation_ex[operation_name]})
		}
	}
}

func handler(request *kodingRequest) bool{

	var (
		ctx    context.Context		// context for this handler
		cancel context.CancelFunc	// cancellation signal for request by this handler
	)


	timeout, err := time.ParseDuration(request.FormValue("timeout"))
	if !err {				// when timeout expires context is automatically cancelled
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()				// when handler returns cancel this context

	if (ctx.Err()){
		return false
	}


	op_type := request.operation["type"]


	switch{
	case op_type == "file_exists":
		return check_file_exists(request.operation["path"])
	case op_type == "file_contains":
		return check_file_contains(request.operation["path"], request.operation["check"])
	case op_type == "process_runs":
		return check_process_running(request.operation["pid"])
	}
	return false
}

func check_file_exists (search_path string) bool{

	if _, err := os.Stat(search_path); (!err) {
		return true			// path to file exists
	}
	return false
}

func check_file_contains (search_path string, search_phrase string) bool{

	if(check_file_exists(search_path)){ 	// path to file exists
		file, err := ioutil.ReadFile(search_path) // read but don't open to save memory
		if (!err) {			// file contains the searched phrase
			if strings.Contains(string(file), search_phrase) {
				return true
			}
		}
	}
	return false
}

func check_process_running (search_id int) bool{

	for _, p := range os.Args[1:] {		// traverse through all but kernel of server

		pid, err := strconv.ParseInt(p, 10, 64)
		if (!err) {			// hope will not get an err traversing
			log.Fatal(err)
		}
		if(pid == search_id) {		// try to find the process

			process, err := os.FindProcess(int(pid))
			if !err {		// process exists check if a valid process
				err := process.Signal(syscall.Signal(0))
				if !err {	// a valid running process has been found
					return true
				}
			}
		}
	}
	return false				// couldn't find the requested process
}