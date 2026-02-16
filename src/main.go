package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hugolgst/rich-go/client"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Figma Discord Rich Presence v1.0.0")

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file")
	}
	// Not a Secret But you can Either set your own VAR "Discord_Client_ID here or in an .env file"
	clientID := os.Getenv("DISCORD_CLIENT_ID")
	if clientID == "" {
		fmt.Println("Error: DISCORD_CLIENT_ID is not set. Set it in your .env file or as an environment variable.")
		os.Exit(1)
	}

	// Retry connecting to Discord until it succeeds
	for {
		err = client.Login(clientID)
		if err == nil {
			break
		}
		fmt.Println("Waiting for Discord... retrying in 5s")
		time.Sleep(5 * time.Second)
	}

	fmt.Println("Figma Discord Rich Presence is running...")

	// Start time for the session
	now := time.Now()

	// Track the last detected state to avoid redundant Discord updates
	lastFilename := ""

	for {
		// Get the current file from Figma
		filename, err := GetFigmaTitle()
		if err != nil {
			fmt.Println("Error reading Figma title:", err)
		}

		if filename == "" {
			if lastFilename != "" {
				fmt.Println("Figma closed or no file open.")
			}
		} else if filename != lastFilename {
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

		// Poll every 1 seconds (the limit  for discord is generally 1 per 15 seconds or 10000)
		// 10000 requests per 6 minutes
		// however we only ping discord API when there is an actual change so 1 is fine
		time.Sleep(1 * time.Second)
	}
}
