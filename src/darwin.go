//go:build darwin

package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const accessibilityCooldown = 15 * time.Second

var accessibilityRetryAfter time.Time

// GetFigmaTitle uses AppleScript to find Figma window titles on macOS.
// NOTE: Requires Accessibility permissions in System Settings -> Privacy & Security -> Accessibility.
func GetFigmaTitle() (string, error) {
	// Avoid hammering osascript if Accessibility is denied.
	if time.Now().Before(accessibilityRetryAfter) {
		return "", nil
	}

	script := `
        tell application "System Events"
            set figmaProcesses to every process whose bundle identifier is "com.figma.Desktop"
            if (count of figmaProcesses) is 0 then
                set figmaProcesses to every process whose name contains "Figma"
            end if

            if (count of figmaProcesses) is 0 then
                return ""
            end if

            set figmaProcess to item 1 of figmaProcesses

            try
                set windowNames to name of every window of figmaProcess
            on error errMsg number errNum
                return "__ERROR__|" & (errNum as text) & "|" & errMsg
            end try

            set oldDelims to AppleScript's text item delimiters
            set AppleScript's text item delimiters to linefeed
            set outputText to windowNames as text
            set AppleScript's text item delimiters to oldDelims
            return outputText
        end tell
    `

	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		outputStr := strings.TrimSpace(string(out))
		if isAccessibilityError(outputStr) {
			accessibilityRetryAfter = time.Now().Add(accessibilityCooldown)
			return "", fmt.Errorf("accessibility permissions required: grant access to figma-rpc in System Settings -> Privacy & Security -> Accessibility")
		}
		if outputStr != "" {
			return "", fmt.Errorf("%w: %s", err, outputStr)
		}
		return "", err
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return "", nil
	}

	if strings.HasPrefix(output, "__ERROR__|") {
		parts := strings.SplitN(output, "|", 3)
		errNum := "unknown"
		errMsg := output
		if len(parts) >= 2 && strings.TrimSpace(parts[1]) != "" {
			errNum = strings.TrimSpace(parts[1])
		}
		if len(parts) == 3 && strings.TrimSpace(parts[2]) != "" {
			errMsg = strings.TrimSpace(parts[2])
		}

		if isAccessibilityError(errMsg) || errNum == "1002" {
			accessibilityRetryAfter = time.Now().Add(accessibilityCooldown)
			return "", fmt.Errorf("accessibility permissions required: grant access to figma-rpc in System Settings -> Privacy & Security -> Accessibility")
		}

		return "", fmt.Errorf("applescript window query failed (%s): %s", errNum, errMsg)
	}

	titles := splitWindowTitles(output)
	homeTitle := false
	var fallbackTitle string

	for _, title := range titles {
		base, ok := trimFigmaSuffix(title)
		if !ok {
			lowerTitle := strings.ToLower(strings.TrimSpace(title))
			if strings.Contains(lowerTitle, "home") || strings.Contains(lowerTitle, "drafts") {
				homeTitle = true
				continue
			}
			if fallbackTitle == "" && strings.TrimSpace(title) != "" {
				fallbackTitle = strings.TrimSpace(title)
			}
			continue
		}

		lowerBase := strings.ToLower(base)
		if lowerBase == "home" || lowerBase == "drafts" {
			homeTitle = true
			continue
		}

		if base != "" {
			return base, nil
		}
	}

	if homeTitle {
		return "Browsing Files", nil
	}

	if fallbackTitle != "" {
		return fallbackTitle, nil
	}

	return "", nil
}

func isAccessibilityError(output string) bool {
	lower := strings.ToLower(output)
	return strings.Contains(lower, "not allowed assistive access") ||
		strings.Contains(lower, "not authorized to send apple events") ||
		strings.Contains(lower, "(-1743)") ||
		strings.Contains(lower, "1002")
}

func splitWindowTitles(output string) []string {
	normalized := strings.ReplaceAll(output, "\r\n", "\n")
	parts := strings.Split(normalized, "\n")
	if len(parts) == 1 {
		parts = strings.Split(output, ", ")
	}

	titles := make([]string, 0, len(parts))
	for _, part := range parts {
		title := strings.TrimSpace(part)
		if title != "" {
			titles = append(titles, title)
		}
	}

	return titles
}

func trimFigmaSuffix(title string) (string, bool) {
	suffixes := []string{
		" - Figma",
		" \u2013 Figma",
		" \u2014 Figma",
	}

	for _, suffix := range suffixes {
		if strings.HasSuffix(title, suffix) {
			return strings.TrimSpace(strings.TrimSuffix(title, suffix)), true
		}
	}

	return "", false
}
