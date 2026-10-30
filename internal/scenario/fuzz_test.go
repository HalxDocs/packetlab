package scenario

import (
	"bytes"
	"os"
	"testing"
)

// FuzzDecode ensures malformed scenario input cannot panic the loader.
func FuzzDecode(f *testing.F) {
	seed, err := os.ReadFile(scenarioPath("two-hosts-one-router.json"))
	if err == nil {
		f.Add(seed)
	}
	f.Add([]byte(validJSON()))
	f.Add([]byte("{"))
	f.Add([]byte(""))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = Decode(bytes.NewReader(data))
	})
}
