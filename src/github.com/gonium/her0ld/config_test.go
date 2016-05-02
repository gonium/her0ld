package her0ld

import (
	"os"
	"reflect"
	"testing"
)

func TestConfigRoundTrip(t *testing.T) {
	filename := "/tmp/her0ld-testsuite-cfg"
	cfg := MkExampleConfig()
	err := SaveConfig(filename, cfg)
	if err != nil {
		t.Fatalf("Failed to save config to file: %s", err.Error())
	}
	// delete file after the test ends.
	defer os.Remove(filename)
	cfg2, err := LoadConfig(filename)
	if err != nil {
		t.Fatalf("Failed to load config from file: %s", err.Error())
	}
	if !reflect.DeepEqual(cfg, cfg2) {
		t.Fatalf("Configurations differ. Pre-save: %#v, Post-save: %#v",
			cfg, cfg2)
	}
}
