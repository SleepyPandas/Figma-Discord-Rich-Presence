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
const appVersion = "2.0.0"

type rpcManagerState struct {
	clientID        string
	rpcEnabled      bool
	privacyMode     bool
	customLabel     string
	connected       bool
	sessionStart    time.Time
	currentFilename string
	lastActivitySig string
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
		clientID:        clientID,
		rpcEnabled:      cfg.RPCEnabled,
		privacyMode:     cfg.PrivacyMode,
		customLabel:     sanitizeCustomLabel(cfg.CustomLabel),
		connected:       false,
		sessionStart:    time.Now(),
		currentFilename: "",
		lastActivitySig: "",
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
			state.lastActivitySig = ""
			if state.connected {
				fmt.Println("Disconnecting from Discord RPC.")
				client.Logout()
				state.connected = false
			}

		case <-events.Reconnect:
			state.rpcEnabled = true
			state.sessionStart = time.Now()
			state.lastActivitySig = ""
			syncActivity(&state, stop, true)

		case updated := <-events.ConfigChanged:
			if updated == nil {
				continue
			}

			prevRPCEnabled := state.rpcEnabled
			prevPrivacyMode := state.privacyMode
			prevLabel := state.customLabel

			state.rpcEnabled = updated.RPCEnabled
			state.privacyMode = updated.PrivacyMode
			state.customLabel = sanitizeCustomLabel(updated.CustomLabel)

			if !state.rpcEnabled {
				state.lastActivitySig = ""
				if state.connected {
					fmt.Println("RPC disabled from settings.")
					client.Logout()
					state.connected = false
				}
				continue
			}

			if !prevRPCEnabled && state.rpcEnabled {
				state.rpcEnabled = true
				state.sessionStart = time.Now()
				state.lastActivitySig = ""
				syncActivity(&state, stop, true)
				continue
			}

			if prevPrivacyMode != state.privacyMode || prevLabel != state.customLabel {
				syncActivity(&state, stop, true)
			}

		case filename := <-filenameUpdates:
			state.currentFilename = filename
			syncActivity(&state, stop, false)
		}
	}
}

func syncActivity(state *rpcManagerState, stop <-chan struct{}, force bool) {
	if !state.rpcEnabled {
		return
	}

	if state.currentFilename == "" {
		if state.connected {
			fmt.Println("Figma closed or no file open. Clearing presence.")
			client.Logout()
			state.connected = false
		}
		state.lastActivitySig = ""
		return
	}

	if !ensureConnected(state, stop) {
		return
	}

	activity := activityFromFilename(state.currentFilename, state.privacyMode, state.customLabel, state.sessionStart)
	signature := activitySignature(activity)
	if !force && signature == state.lastActivitySig {
		return
	}

	fmt.Println("State changed:", state.currentFilename)
	if err := client.SetActivity(activity); err != nil {
		fmt.Println("Failed to set activity:", err)
		return
	}

	state.lastActivitySig = signature
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

func activityFromFilename(filename string, privacyMode bool, customLabel string, start time.Time) client.Activity {
	if privacyMode {
		state := sanitizeCustomLabel(customLabel)
		details := "Editing File"
		smallImage := "edit"
		smallText := "Editing"

		if filename == "Browsing Files" {
			details = "In Home"
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

func activitySignature(activity client.Activity) string {
	return fmt.Sprintf("%s|%s|%s|%s", activity.Details, activity.State, activity.SmallImage, activity.SmallText)
}

func sanitizeCustomLabel(label string) string {
	if label == "" {
		return "Working on a project"
	}
	return label
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
