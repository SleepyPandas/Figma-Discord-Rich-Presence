//go:build darwin

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetFigmaTitle uses AppleScript to find the Figma window title on macOS.
// NOTE: Requires Accessibility permissions in System Settings → Privacy & Security → Accessibility.
func GetFigmaTitle() (string, error) {
	// AppleScript to get the name of the frontmost Figma window
	script := `
		tell application "System Events"
			if exists (process "Figma") then
				tell process "Figma"
					set windowList to name of every window
				end tell
				return windowList as text
			else
				return ""
			end if
		end tell
	`

	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		outputStr := strings.TrimSpace(string(out))
		// Detect accessibility permission errors
		if strings.Contains(outputStr, "not allowed assistive access") ||
			strings.Contains(outputStr, "1002") {
			return "", fmt.Errorf("accessibility permissions required: grant access to figma-rpc in System Settings → Privacy & Security → Accessibility")
		}
		return "", err
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return "", nil
	}

	// AppleScript returns comma-separated window names if multiple
	// We iterate through them with the same priority logic as Windows
	windows := strings.Split(output, ", ")

	var fileTitle string
	var homeTitle string

	for _, title := range windows {
		if strings.Contains(title, "Figma") {
			// Priority 1: An actual file (e.g., "Project - Figma")
			if strings.Contains(title, " - Figma") && !strings.Contains(title, "Home") {
				fileTitle = strings.TrimSuffix(title, " - Figma")
				break // Found a file, stop searching
			}

			// Priority 2: Home screen
			if strings.Contains(title, "Home - Figma") || strings.Contains(title, "Drafts - Figma") {
				homeTitle = "Browsing Files"
			}
		}
	}

	if fileTitle != "" {
		return fileTitle, nil
	}
	return homeTitle, nil
}
