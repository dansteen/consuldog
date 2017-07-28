package datadog

import (
	"log"
	"os"
	"time"

	ps "github.com/mitchellh/go-ps.git"
	"github.com/spf13/viper"
)

// we need to implement a signal for SigHup
type SigHup struct {
}

func (sig SigHup) String() string { return "HUP" }
func (sig SigHup) Signal()        {}

// Reloader will reload the datadog process when a value is set on the reload channel
func Reloader(reloadRequested <-chan bool, stop <-chan bool) {
	// log errors to stderr
	logger := log.New(os.Stderr, log.Prefix(), 0)
	// set up a ticker to trigger the actual reload
	ticker := time.NewTicker(time.Duration(viper.GetInt64("datadogMinReloadInterval")) * time.Second)
	// store a value to see if we should actually reload or not
	reload := false
	// get the Effective UID of the current consuldog process
	// not used since we don't get the UID of the other proc.  See todo note below.
	//ourUID := os.Geteuid()
	// get the name of the process we are looking for
	datadogProcName := viper.GetString("datadogProcName")

	// listen for requests
	for {
		select {
		case <-ticker.C:
			// we only proceed if a reload has been requested
			if reload == true {
				// get a process list
				processes, err := ps.Processes()
				if err != nil {
					logger.Println("Could not get process list.  Skipping datadog reload.")
					continue
				}
				// let us know if we reloaded something
				reloaded := false
				// find a process that is owned by the same uid that we are using, and is named datadogProcName
				for _, process := range processes {
					pid := process.Pid()
					name := process.Executable()
					osProcess, err := os.FindProcess(pid)
					// if we couldn't find the process we just go on to the next one.  There are plenty of transient processes around
					if err != nil {
						continue
					}
					if name == datadogProcName {
						// TODO: this is messy since we just send HUP signals to every process correctly named and ignore the failures unless
						// we have no successes.   We should really search for processes owned by our UID, but there is no standard cross-platform
						// way to do that.  For now this will work fine in almost all situations.
						err := osProcess.Signal(SigHup{})
						if err == nil {
							reloaded = true
						}
					}
				}
				// see if we successfully reloaded anything
				if reloaded == false {
					logger.Printf("Could not find a process named %s that we could signal (The process needs to be running the same UID as this one). Skipping Datadog reload.\n", datadogProcName)
				} else {
					log.Printf("%s Reloaded.\n", datadogProcName)
				}
				// reset our reload value
				reload = false
			}
		case <-reloadRequested:
			reload = true
		case <-stop:
			return
		}
	}
}
