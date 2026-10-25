package api

import (
	"regexp"
	"strings"
	"testing"
)

// The Scalar bundle runs on the API origin, so it must be pinned and integrity checked.
func TestReferencePageScriptHasSRI(t *testing.T) {
	if !strings.Contains(referencePage, `integrity="sha384-`) {
		t.Fatal("the Scalar script has no integrity attribute")
	}
	if !strings.Contains(referencePage, `crossorigin="anonymous"`) {
		t.Fatal("the Scalar script needs crossorigin for SRI to apply")
	}
	if !regexp.MustCompile(`api-reference@\d+\.\d+\.\d+/dist/browser/standalone\.js`).MatchString(referencePage) {
		t.Fatal("the Scalar script must name an exact version and its dist file")
	}
}
