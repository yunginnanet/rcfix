package main

import (
	"bytes"
	"testing"

	"compress/gzip"
	"encoding/base64"
)

// FallenAngel: I can fix her
const FallenAngel = `H4sIAAAAAAACA3VWbU/jRhD+7l8xUj4UJBJTdCedUF0JSA6iXgGRUFShKlrbY7xivevbXSfk33dm1w6UpFIUW/syL88884yfH7X0/yRTdIWVrZdGZ9fSo4AjesBG+hoEFF0LpgJaPk4uKo820+g3xr5OvLAv6JPRaJSMYGr0Lx4qw0vgDYiyBF8jlMKLXDgEh3YtC1rAFnWJupDo4t1k9CS0d1mzdT/VpD+XjKKzT4u7s8JKUeZ7pz8vD+db4/yLxQMODu3svGBTiKLGcs/P/sZwx2Ip3efznxZ7zOYVdE7qF3CmeCXYROHlWnAdGEhohNRQe9+mO6CitRcu0oS3J/FmMnrAn52kLA7t9d6O/jYdFEKDUM5Aa81algjhPIjoCCqhVC6KV1ooUwrBuXoIzhtzzLbgQgO+iaZVyLxI0Rep2zqPTdk/0wMxcAT8g+fIOXrbp90T5rAYLsC9sP6u6hPaQReNxFPBzCLu7J2DH5KC0QtvUTTZb840uGqN9b/z3q2ZohLbzNtuZ3Ounaf0g1EuJZaX2yzG7965HlJJnnuvdBgedWGaBrUPfNf45kFJjSAr2BLitVgjWCSaxY5SxjsGrpIKHeMM3DECbpbLe/h6egpoLSGfYyE6FyD2tSDPlE8j/e3d9/mPWfb17MvZt2/n8ZE8IEVu/QKL7Mwly22LmZNcoOTRRbok19Z0bXh7otYlzk2JLYU3dpula2FTJfNYtvQjL6lWb339z0kBlHLQl5pbvCBkPYasU9vpeJ+Yq0q0J7CpZVFTykpBYbRnLvPJvkxkMyDAvGyE7gj4bTTIfj+YKw06VhZWnJMdKixPplMlaOOhReuo1tQ+1jhHWOeGMCayPnTaywbfUw0mk9kbFgsGLEs7R6mbQqg0l4PHDbFwPKaYK/kS6R2BEW07kVoOaGdCbcTWJTO9ltZoJkD2uJg9sBO4uftzlqU1cY4vw/V8ObtYPd09/LGazh8+IR4BZ6rIyEBg9SV8yyFualasqBKGALRMB00qWolOebi/WN7AUcS6pghRk1YE9Rh6VP7XdrSxZsiMDkRkb46kYOzNmJ+BpmSEFMtbmXdBj/oLxyfQ7eieozKbyHWmcVR84YdoKQOO7sBsaJUoAhPGqnJAyAvK0UR+hISi5T7yjdDhFmqRB8JwwD++L8B1LTd0MvpYAr6fpa3wdeoNA8yVPY9/LvxzzXcvvPRegMEVLZYQNYlHGbCbPqhfT8++fACB7nLUNBBhLVRHHR2OnQBVgIm6r+ycvyCaRvOSxICtkxLQLzdrjFp5JVqRSyX99tJ0uqSmWKDPri7uV7ez5epyfjtdEdf+ml/NaCg0uaRQdldosP7PyTh1NLAY0ihw6LhPUfftftBpqO0BF2CCdJ8E4BqxHcAjDw6jFpIUEBZsOWDDqdOAceFF8BbF8MEgA09UalslsWQ+xH6kSVWgcxNYfjRJbqJRYomLtGIxB0fh5uaNzMXbw2cHZ0HEXDNxhd4OVulcZU0DJDiaAw0S3Vq5Jm164Q8U6HWrps8ErqgFLRp0LTF4EgG9t1xaZKl1WcivnxG7gbKbJg21rByzlWGg/AumuTXcgwkAAA==`

var TestService []byte

func init() {
	raw, _ := base64.StdEncoding.DecodeString(FallenAngel)
	gzr, _ := gzip.NewReader(bytes.NewReader(raw))
	b := bytes.Buffer{}
	_, _ = b.ReadFrom(gzr)
	gzr.Close()
	TestService = b.Bytes()
}

func TestReadSystemdService(t *testing.T) {
	service, err := ReadBrokenService(bytes.NewReader(TestService))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedName := "gitea"
	if service.Name != expectedName {
		t.Errorf("expected Name %q, got %q", expectedName, service.Name)
	}

	expectedDescription := "Gitea (Git with a cup of tea)"
	if service.Description != expectedDescription {
		t.Errorf("expected Description %q, got %q", expectedDescription, service.Description)
	}

	expectedExecStart := "/usr/local/bin/gitea web --config /etc/gitea/app.ini"
	if service.ExecStart != expectedExecStart {
		t.Errorf("expected ExecStart %q, got %q", expectedExecStart, service.ExecStart)
	}

	expectedExecStop := ""
	if service.ExecStop != expectedExecStop {
		t.Errorf("expected ExecStop %q, got %q", expectedExecStop, service.ExecStop)
	}

}
