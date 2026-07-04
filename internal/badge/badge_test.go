package badge

import (
	"strings"
	"testing"
)

func TestSVGContainsLabelAndMessage(t *testing.T) {
	svg := SVG("pipeline", "42", ColorBrightGreen)
	if !strings.Contains(svg, ">pipeline<") {
		t.Error("expected label in SVG")
	}
	if !strings.Contains(svg, ">42<") {
		t.Error("expected message in SVG")
	}
	if !strings.Contains(svg, ColorBrightGreen) {
		t.Error("expected color in SVG")
	}
	if !strings.Contains(svg, `x="310"`) {
		t.Error("expected label text x coordinate in pre-scale units")
	}
}

func TestSVGEscapesXML(t *testing.T) {
	svg := SVG(`a&b`, `<tag>`, ColorBlue)
	if strings.Contains(svg, `a&b`) {
		t.Error("expected ampersand to be escaped")
	}
	if strings.Contains(svg, "<tag>") {
		t.Error("expected angle brackets to be escaped")
	}
}
