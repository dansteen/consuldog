package datadog

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

// Uid is a struct for the Uid/gid lines of proc/*/status
type Uid struct {
	Real       int
	Effective  int
	Saved      int
	Filesystem int
}

// Status is a struct for the content of proc/*/status
// (only the parts we care about
type Status struct {
	Name string
	Pid  int
	Uid  Uid
}

//UnmarshalText will turn the status file into a usable structure for us
func (status *Status) UnmarshalText(text []byte) error {
	// scan our data and set our values
	textBuf := bytes.NewBuffer(text)
	scanner := bufio.NewScanner(textBuf)
	for scanner.Scan() {
		// the data is : delimited
		line := strings.SplitN(scanner.Text(), ":", 2)
		if line[0] == "Name" {
			status.Name = strings.TrimSpace(line[1])
		} else if line[0] == "Pid" {
			status.Pid, _ = strconv.Atoi(strings.TrimSpace(line[1]))
		} else if line[0] == "Uid" {
			// break up our line
			uidLine := strings.SplitN(line[1], "\t", 4)
			real, _ := strconv.Atoi(uidLine[0])
			eff, _ := strconv.Atoi(uidLine[1])
			saved, _ := strconv.Atoi(uidLine[2])
			file, _ := strconv.Atoi(uidLine[3])
			status.Uid = Uid{
				Real:       real,
				Effective:  eff,
				Saved:      saved,
				Filesystem: file,
			}
		}
	}

	// if there were any errors we return them
	return scanner.Err()
}

// Reloader will reload the datadog process when a value is set on the reload channel
func Reloader(reloadRequested <-chan bool, stop <-chan bool) {
	// log errors to stderr
	logger := log.New(os.Stderr, log.Prefix(), 0)
	// set up a ticker to trigger the actual reload
	ticker := time.NewTicker(time.Duration(viper.GetInt64("datadogMinReloadInterval")) * time.Second)
	// store a value to see if we should actually reload or not
	reload := false
	// get the Effective UID of the current consuldog process
	ourUID := os.Geteuid()
	// get the name of the process we are looking for
	datadogProcName := viper.GetString("datadogProcName")

	// listen for requests
	for {
		select {
		case <-ticker.C:
			// we only proceed if a reload has been requested
			if reload == true {
				// record if we have actually reloaded anything
				reloaded := false
				// place to store our /proc/*/status information
				statusData := make([]byte, 2048)
				status := Status{}
				// run through all the processes
				paths, _ := filepath.Glob("/proc/*/status")
				for _, path := range paths {
					// see if this one matches the name we are looking for and the user
					// we ignore errors for all this since we expect that some proceses will disappear prior to us being done with them
					// if the name of the process is greater than 512 bytes we don't bother to continue reading as it is unlikely
					// that we have the file that we are looking for
					commFile, _ := os.Open(path)
					commFile.Read(statusData)
					// convert the file to our struct
					status.UnmarshalText(statusData)
					// once we have our data we see if it matches
					if status.Name == datadogProcName && status.Uid.Effective == ourUID {
						// grab the process and send a signal
						osProcess, _ := os.FindProcess(status.Pid)
						err := osProcess.Signal(syscall.SIGHUP)
						// if we succeeded we make a note of it, otherwise we print a message
						if err != nil {
							logger.Printf("Failed to send reload signal to %s (%v):\n", status.Name, status.Pid)
							logger.Println(err)
						} else {
							log.Printf("Reloaded %s (%v)\n", status.Name, status.Pid)
							reloaded = true
						}
					}
				}

				// if we haven't actually reloaded anything we post a message
				if reloaded == false {
					// convert our uid to a name if we can
					var userName string
					ourUser, err := user.LookupId(strconv.Itoa(ourUID))
					if err != nil {
						userName = "<unknown>"
					} else {
						userName = ourUser.Username
					}
					logger.Printf("Could not find or successfully signal any processes named '%s' owned  by %s(%d). Datadog Reload Skipped.\n", datadogProcName, userName, ourUID)
					// reset our reload value
				}
				reload = false
			}
		case <-reloadRequested:
			reload = true
		case <-stop:
			return
		}
	}
}
