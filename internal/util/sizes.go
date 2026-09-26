package util

// MinWidth and MinHeight are the smallest sizes the window is allowed to be.
// Kept here so both main.go and window.go agree on the same numbers.
const (
	MinW = 720
	MinH = 480

	// Version is the app version, shown in the UI and used for the About
	// display. Update this when tagging a new release.
	Version = "1.0.0"

	// Author is the display name shown in the header.
	Author = "WAseem KHan"

	// GitHubURL is the author's profile link.
	GitHubURL = "https://github.com/khpalwatan"
)

// MinWidth returns the minimum window width.
func MinWidth() int { return MinW }

// MinHeight returns the minimum window height.
func MinHeight() int { return MinH }