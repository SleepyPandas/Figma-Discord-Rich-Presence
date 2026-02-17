package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hugolgst/rich-go/client"
)

// Replace with your Own Public Key if you desire

const discordClientID = "1473014472498086092"
const appVersion = "1.0.1"

func main() {
	fmt.Printf("Figma Discord Rich Presence v%s\n", appVersion)

	clientID := discordClientID

	// Retry connecting to Discord until it succeeds
	var err error
	for {
		err = client.Login(clientID)
		if err == nil {
			break
		}
		fmt.Println("Waiting for Discord... retrying in 5s")
		time.Sleep(5 * time.Second)
	}

	// Handle graceful shutdown (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down... logging out of Discord.")
		client.Logout()
		os.Exit(0)
	}()

	fmt.Println("Figma Discord Rich Presence is running...")

	// Start time for the session
	now := time.Now()

	// Track the last detected state to avoid redundant Discord updates
	lastFilename := ""

	// Track whether we are currently connected to Discord
	loggedIn := true
	lastReadErr := ""
	lastReadErrAt := time.Time{}
	emptyPolls := 0

	for {
		// Get the current file from Figma
		filename, err := GetFigmaTitle()
		if err != nil {
			errMsg := err.Error()
			nowErr := time.Now()
			if errMsg != lastReadErr || nowErr.Sub(lastReadErrAt) >= 30*time.Second {
				fmt.Println("Error reading Figma title:", err)
				lastReadErr = errMsg
				lastReadErrAt = nowErr
			}

			time.Sleep(1 * time.Second)
			continue
		}
		lastReadErr = ""
		lastReadErrAt = time.Time{}

		if filename == "" {
			emptyPolls++
			if emptyPolls == 1 && lastFilename != "" {
				fmt.Println("Figma title temporarily unavailable, waiting for confirmation...")
			}

			// Require several consecutive empty polls to avoid flapping on transient AppleScript misses.
			if emptyPolls < 3 {
				time.Sleep(1 * time.Second)
				continue
			}

			// Figma was open before but is now closed clear the presence
			if lastFilename != "" {
				fmt.Println("Figma closed or no file open. Clearing presence.")
				client.Logout()
				loggedIn = false
				lastFilename = ""
			}
		} else {
			if emptyPolls > 0 && lastFilename != "" {
				fmt.Println("Figma title recovered.")
			}
			emptyPolls = 0

			// Figma is open reconnect to Discord if we logged out
			if !loggedIn {
				fmt.Println("Figma detected again, reconnecting to Discord...")
				for {
					err = client.Login(clientID)
					if err == nil {
						break
					}
					fmt.Println("Waiting for Discord... retrying in 5s")
					time.Sleep(5 * time.Second)
				}
				loggedIn = true
				now = time.Now() // reset session timer
			}

			if filename != lastFilename {
				fmt.Println("State changed:", filename)

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

				err = client.SetActivity(client.Activity{
					State:      state,
					Details:    details,
					LargeImage: "largeimageid",
					LargeText:  "Figma",
					SmallImage: smallImage,
					SmallText:  smallText,
					Timestamps: &client.Timestamps{
						Start: &now,
					},
				})

				if err != nil {
					fmt.Println("Failed to set activity:", err)
				}
			}

			lastFilename = filename
		}

		// Poll every 1 seconds (the limit  for discord is generally 1 per 15 seconds or 10000)
		// 10000 requests per 6 minutes
		// however we only ping discord API when there is an actual change so 1 is fine
		time.Sleep(1 * time.Second)
	}
}
