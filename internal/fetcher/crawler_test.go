package fetcher

import (
	"reflect"
	"testing"
)

func TestFilterLinks(t *testing.T) {
	var filterTest = []struct {
		name  string
		links []string
		want  []string
	}{
		{
			name:  "absolete links included",
			links: []string{"https://example.com", "script1.js", "/page2.html"},
			want:  []string{"https://example.com", "https://example.com/script1.js", "https://example.com/page2.html"},
		},
		{
			name:  "external links included",
			links: []string{"https://example.com", "script1.js", "/page2.html", "http://google.com", "https://example1.com/script2.js", "http://example1.com/insecure.html"},
			want:  []string{"https://example.com", "https://example.com/script1.js", "https://example.com/page2.html"},
		},
	}

	c := Crawler{}
	for _, tt := range filterTest {
		t.Run(tt.name, func(t *testing.T) {
			got := c.filterLinks("https://example.com", tt.links)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
