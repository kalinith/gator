package config
import "testing"

func TestOpenConfig(t *testing.T) {
	empty_config := Config{}
	expected := Config{"postgres://example",""}
	actual, error := Read()
	//was there an error
	if error != nil {
		t.Errorf("failed to access the config file. Error: %s", error)
		return
	}
	if actual == empty_config {
		t.Errorf("Read() returned an empty Config struct when it should contain data")
	}

	if actual.DatabaseURL != expected.DatabaseURL {
		t.Errorf("the database URL:%s does not match the expected:%s", actual, expected)
	}

}
