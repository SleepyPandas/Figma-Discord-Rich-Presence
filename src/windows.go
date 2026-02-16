//go:build windows

package main

import (
	"strings"
	"syscall"
	"unsafe"
)

var (
	// Load the Windows User32 DLL (contains window management functions)
	user32             = syscall.NewLazyDLL("user32.dll")
	procEnumWindows    = user32.NewProc("EnumWindows")
	procGetWindowTextW = user32.NewProc("GetWindowTextW")
)

// GetFigmaTitle searches all open windows for Figma and returns the active file name.
func GetFigmaTitle() (string, error) {
	var fileTitle string
	var homeTitle string

	// Define the callback function that Windows will call for every window it finds
	cb := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		// 1. Create a buffer to hold the title (256 chars is usually enough)
		buf := make([]uint16, 256)

		// 2. Read the window title into the buffer
		ret, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))

		if ret > 0 {
			// Convert UTF-16 buffer to a Go string
			title := syscall.UTF16ToString(buf)

			// 3. Filter for Figma windows
			if strings.Contains(title, "Figma") {

				// Priority 1: Check for an actual file (e.g., "Project – Figma")
				// Note: Figma often uses a specific dash " – ". Copy-paste yours if this fails!
				if strings.Contains(title, " - Figma") && !strings.Contains(title, "Home") {
					// Clean up the string to get just the file name
					fileTitle = strings.TrimSuffix(title, " - Figma")
					return 0 // Return 0 to STOP searching (we found it!)
				}

				// Priority 2: Check for the Home screen
				if strings.Contains(title, "Home - Figma") || strings.Contains(title, "Drafts - Figma") {
					homeTitle = "Browsing Files"
				}
			}
		}
		return 1 // Return 1 to CONTINUE searching
	})

	// Start the enumeration process
	procEnumWindows.Call(cb, 0)

	// Return the best match we found
	if fileTitle != "" {
		return fileTitle, nil
	}
	return homeTitle, nil
}
