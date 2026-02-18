package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// websiteDarkTheme matches the website dark mode: monochrome foreground and
// near-black surfaces with subtle light borders.
type websiteDarkTheme struct {
	fallback fyne.Theme
}

func newWebsiteDarkTheme() fyne.Theme {
	return &websiteDarkTheme{fallback: theme.DefaultTheme()}
}

func (t *websiteDarkTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255} // #000000
	case theme.ColorNameForeground:
		return color.NRGBA{R: 245, G: 245, B: 245, A: 255} // #f5f5f5
	case theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		return color.NRGBA{R: 163, G: 163, B: 163, A: 255} // #a3a3a3
	case theme.ColorNameButton, theme.ColorNameInputBackground, theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 10, G: 10, B: 10, A: 209} // rgba(10,10,10,0.82)
	case theme.ColorNameMenuBackground:
		return color.NRGBA{R: 10, G: 10, B: 10, A: 230}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 10, G: 10, B: 10, A: 220}
	case theme.ColorNameInputBorder, theme.ColorNameSeparator:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 36} // rgba(255,255,255,0.14)
	case theme.ColorNameFocus:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 92} // rgba(255,255,255,0.36)
	case theme.ColorNameHover:
		return color.NRGBA{R: 30, G: 30, B: 30, A: 242} // rgba(30,30,30,0.95)
	case theme.ColorNamePressed:
		return color.NRGBA{R: 34, G: 34, B: 34, A: 250} // rgba(34,34,34,0.98)
	case theme.ColorNamePrimary, theme.ColorNameHyperlink:
		return color.NRGBA{R: 245, G: 245, B: 245, A: 255}
	case theme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 17, G: 17, B: 17, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 20, G: 20, B: 20, A: 200}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 245, G: 245, B: 245, A: 46}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 89}
	case theme.ColorNameScrollBarBackground:
		return color.NRGBA{R: 10, G: 10, B: 10, A: 255}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 200}
	}

	// Force fallback lookups to dark variant so this theme stays dark-only.
	return t.fallback.Color(name, theme.VariantDark)
}

func (t *websiteDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.fallback.Font(style)
}

func (t *websiteDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.fallback.Icon(name)
}

func (t *websiteDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.fallback.Size(name)
}
