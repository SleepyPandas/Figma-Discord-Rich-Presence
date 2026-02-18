package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hugolgst/rich-go/client"
)

// Replace with your own public key if desired.
const discordClientID = "1473014472498086092"
const appVersion = "1.0.1"

type rpcManagerState struct {
	clientID     string
	rpcEnabled   bool
	connected    bool
	sessionStart time.Time
	lastFilename string
}

func main() {
	fmt.Printf("Figma Discord Rich Presence v%s\n", appVersion)

	cfg, err := LoadConfig()
	if err != nil {
		fmt.Println("Warning: loading config had issues:", err)
	}

	events := NewUIEvents()
	ui := SetupUI(cfg, events)

	stop := make(chan struct{})
	filenameUpdates := make(chan string, 1)
	var wg sync.WaitGroup

	wg.Add(1)
	go runFigmaPoller(filenameUpdates, stop, &wg)

	wg.Add(1)
	go runRPCManager(discordClientID, cfg, events, filenameUpdates, stop, &wg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		ui.App.Quit()
	}()

	fmt.Println("Figma Discord Rich Presence is running...")
	ui.Run()

	close(stop)
	wg.Wait()
	fmt.Println("Exited cleanly.")
}

func runFigmaPoller(filenameUpdates chan string, stop <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	lastReadErr := ""
	lastReadErrAt := time.Time{}
	emptyPolls := 0
	lastSent := ""
	hasLastSent := false

	for {
		select {
		case <-stop:
			return
		default:
		}

		filename, err := GetFigmaTitle()
		if err != nil {
			errMsg := err.Error()
			nowErr := time.Now()
			if errMsg != lastReadErr || nowErr.Sub(lastReadErrAt) >= 30*time.Second {
				fmt.Println("Error reading Figma title:", err)
				lastReadErr = errMsg
				lastReadErrAt = nowErr
			}
			if !sleepWithStop(1*time.Second, stop) {
				return
			}
			continue
		}
		lastReadErr = ""
		lastReadErrAt = time.Time{}

		if filename == "" {
			emptyPolls++
			if emptyPolls == 1 && hasLastSent && lastSent != "" {
				fmt.Println("Figma title temporarily unavailable, waiting for confirmation...")
			}
			if emptyPolls < 3 {
				if !sleepWithStop(1*time.Second, stop) {
					return
				}
				continue
			}
		} else {
			if emptyPolls > 0 && hasLastSent && lastSent != "" {
				fmt.Println("Figma title recovered.")
			}
			emptyPolls = 0
		}

		if !hasLastSent || filename != lastSent {
			pushLatestFilename(filenameUpdates, filename)
			lastSent = filename
			hasLastSent = true
		}

		if !sleepWithStop(1*time.Second, stop) {
			return
		}
	}
}

func runRPCManager(clientID string, cfg *Config, events *UIEvents, filenameUpdates <-chan string, stop <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	state := rpcManagerState{
		clientID:     clientID,
		rpcEnabled:   cfg.RPCEnabled,
		connected:    false,
		sessionStart: time.Now(),
		lastFilename: "",
	}

	if !state.rpcEnabled {
		fmt.Println("RPC is disabled in settings. Waiting for Reconnect.")
	}

	for {
		select {
		case <-stop:
			if state.connected {
				client.Logout()
			}
			return

		case <-events.Disconnect:
			state.rpcEnabled = false
			state.lastFilename = ""
			if state.connected {
				fmt.Println("Disconnecting from Discord RPC.")
				client.Logout()
				state.connected = false
			}

		case <-events.Reconnect:
			state.rpcEnabled = true
			state.sessionStart = time.Now()
			state.lastFilename = ""
			ensureConnected(&state, stop)

		case updated := <-events.ConfigChanged:
			if updated == nil {
				continue
			}
			if updated.RPCEnabled == state.rpcEnabled {
				continue
			}
			if updated.RPCEnabled {
				state.rpcEnabled = true
				state.sessionStart = time.Now()
				state.lastFilename = ""
				ensureConnected(&state, stop)
			} else {
				state.rpcEnabled = false
				state.lastFilename = ""
				if state.connected {
					fmt.Println("RPC disabled from settings.")
					client.Logout()
					state.connected = false
				}
			}

		case filename := <-filenameUpdates:
			handleFilenameUpdate(&state, filename, stop)
		}
	}
}

func handleFilenameUpdate(state *rpcManagerState, filename string, stop <-chan struct{}) {
	if !state.rpcEnabled {
		return
	}

	if filename == "" {
		if state.lastFilename != "" {
			fmt.Println("Figma closed or no file open. Clearing presence.")
		}
		if state.connected {
			client.Logout()
			state.connected = false
		}
		state.lastFilename = ""
		return
	}

	if !ensureConnected(state, stop) {
		return
	}

	if filename == state.lastFilename {
		return
	}

	fmt.Println("State changed:", filename)
	if err := client.SetActivity(activityFromFilename(filename, state.sessionStart)); err != nil {
		fmt.Println("Failed to set activity:", err)
		return
	}

	state.lastFilename = filename
}

func ensureConnected(state *rpcManagerState, stop <-chan struct{}) bool {
	if !state.rpcEnabled {
		return false
	}
	if state.connected {
		return true
	}

	for {
		err := client.Login(state.clientID)
		if err == nil {
			state.connected = true
			return true
		}
		fmt.Println("Waiting for Discord... retrying in 5s")
		select {
		case <-stop:
			return false
		case <-time.After(5 * time.Second):
		}
	}
}

func activityFromFilename(filename string, start time.Time) client.Activity {
	details := "Editing File"
	state := filename
	smallImage := "edit"
	smallText := "Editing"

	if filename == "Browsing Files" {
		details = "In Home"
		state = "Browsing Files"
		smallImage = "folder"
		smallText = "Browsing"
	}

	return client.Activity{
		State:      state,
		Details:    details,
		LargeImage: "largeimageid",
		LargeText:  "Figma",
		SmallImage: smallImage,
		SmallText:  smallText,
		Timestamps: &client.Timestamps{
			Start: &start,
		},
	}
}

func pushLatestFilename(ch chan string, value string) {
	select {
	case ch <- value:
	default:
		select {
		case <-ch:
		default:
		}
		ch <- value
	}
}

func sleepWithStop(duration time.Duration, stop <-chan struct{}) bool {
	select {
	case <-stop:
		return false
	case <-time.After(duration):
		return true
	}
}
