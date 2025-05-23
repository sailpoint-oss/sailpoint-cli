package rule

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestRuleXMLSerialization(t *testing.T) {
	// Define the Rule struct with characters that should remain unchanged
	rule := Rule{
		Name: "Test Rule",
		Type: "TestType",
		Description: struct {
			Content string `xml:",innerxml"`
		}{
			Content: `This contains "quotes", <tags>, and & special characters. * '`,
		},
		Source: struct {
			Content string `xml:",cdata"`
		}{
			Content: `if (x > '5') { return true; }`, // Should be wrapped in CDATA
		},
	}

	// Marshal to XML
	xmlBytes, err := xml.MarshalIndent(rule, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal XML: %v", err)
	}

	xmlString := string(xmlBytes)

	// Ensure special characters remain as-is (not encoded)
	tests := []string{
		`"quotes"`,  // Should not be converted to &quot;
		`<tags>`,    // Should not be converted to &lt; or &gt;
		`& special`, // Should not be converted to &amp;
		`<![CDATA[if (x > '5') { return true; }]]>`, // CDATA should remain unchanged
		`*`, // Should remain as-is
		`'`, // Should remain as-is
	}

	// Loop through test cases
	t.Log("Checking for expected string in XML output:", xmlString)

	for _, expected := range tests {
		if !strings.Contains(xmlString, expected) {
			t.Errorf("Expected XML to contain %q but it was missing.\nActual XML:\n%s", expected, xmlString)
		}
	}
}
