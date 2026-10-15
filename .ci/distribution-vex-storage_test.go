package storage

import (
	"strings"
	"testing"
)

func TestHarborVEXStoredContentBound(t *testing.T) {
	for _, size := range []int{int(maxBlobGetSize), int(maxBlobGetSize) + 1} {
		content, err := readAllLimited(strings.NewReader(strings.Repeat("x", size)), maxBlobGetSize)
		if (err != nil) != (size > int(maxBlobGetSize)) || len(content) > int(maxBlobGetSize) {
			t.Fatalf("size=%d returned=%d error=%v", size, len(content), err)
		}
	}
}
