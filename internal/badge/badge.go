package badge

import (
	"fmt"
	"strings"
)

// Color constants matching common shields.io badge colors.
const (
	ColorBrightGreen = "#4c1"
	ColorGreen       = "#97ca00"
	ColorYellow      = "#dfb317"
	ColorOrange      = "#fe7d37"
	ColorRed         = "#e05d44"
	ColorLightGrey   = "#9f9f9f"
	ColorBlue        = "#007ec6"
)

const charWidth = 6.5
const badgePadding = 10

// SVG renders a shields.io-style flat badge.
func SVG(label, message, color string) string {
	label = escapeXML(label)
	message = escapeXML(message)

	labelWidth := textWidth(label)
	messageWidth := textWidth(message)
	width := labelWidth + messageWidth

	labelCenter := labelWidth / 2
	messageCenter := labelWidth + messageWidth/2

	// Text coordinates are in pre-scale units (x10) inside a scale(.1) group,
	// matching the shields.io badge format.
	labelX := labelCenter * 10
	messageX := messageCenter * 10
	labelTextLen := int(float64(len(label)) * charWidth * 10)
	messageTextLen := int(float64(len(message)) * charWidth * 10)

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%d" height="20" role="img" aria-label="%s: %s"><title>%s: %s</title><linearGradient id="s" x2="0" y2="100%%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="%d" height="20" fill="#555"/><rect x="%d" width="%d" height="20" fill="%s"/><rect width="%d" height="20" fill="url(#s)"/><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><g transform="scale(.1)"><text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" textLength="%d">%s</text><text x="%d" y="140" textLength="%d">%s</text></g><g transform="scale(.1)"><text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" textLength="%d">%s</text><text x="%d" y="140" textLength="%d">%s</text></g></g></g></svg>`,
		width, label, message, label, message,
		width,
		labelWidth, labelWidth, messageWidth, color, width,
		labelX, labelTextLen, label, labelX, labelTextLen, label,
		messageX, messageTextLen, message, messageX, messageTextLen, message,
	)
}

func textWidth(s string) int {
	return int(float64(len(s))*charWidth) + badgePadding
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
