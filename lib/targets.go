package vegeta

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
)

// Target is a HTTP request blueprint
type Target struct {
	Method string
	URL    string
	Body   []byte
	Header http.Header
	Size   int64
}

// Request creates an *http.Request out of Target and returns it along with an
// error in case of failure.
func (t *Target) Request() (*http.Request, error) {
	req, err := http.NewRequest(t.Method, t.URL, bytes.NewBuffer(t.Body))
	if err != nil {
		return nil, err
	}
	for k, vs := range t.Header {
		req.Header[k] = make([]string, len(vs))
		copy(req.Header[k], vs)
	}
	if host := req.Header.Get("Host"); host != "" {
		req.Host = host
	}
	return req, nil
}

// Targets is a slice of Targets which can be shuffled
type Targets []Target

// NewTargetsFrom reads targets out of a line separated source skipping empty lines
// It sets the passed body and http.Header on all targets.
func NewTargetsFrom(source io.Reader, body []byte, header http.Header, ranges string) (Targets, error) {
	scanner := bufio.NewScanner(source)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()

		if line = strings.TrimSpace(line); line != "" && line[0:2] != "//" {
			// Skipping comments or blank lines
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return NewTargets(lines, body, header, ranges)
}

// NewTargets instantiates Targets from a slice of strings.
// It sets the passed body and http.Header on all targets.
func NewTargets(lines []string, body []byte, header http.Header, ranges string) (Targets, error) {
	var targets Targets
	for _, line := range lines {
		ps := strings.Split(line, " ")
		if len(ps) != 2 {
			return nil, fmt.Errorf("invalid request format: `%s`", line)
		}
		// size := 0
		// if ranges != "off" {
		// 	// HEAD request, capture content length
		// 	resp, err := http.Head(ps[1])
		// 	if err != nil {
		// 		return nil, fmt.Errorf("couldn't grab content-length: `%s`", ps[1])
		// 	}

		// 	return nil, fmt.Errorf("couldn't grab content-length: `%s`", resp)
		// }
		targets = append(targets, Target{Method: ps[0], URL: ps[1], Body: body, Header: header})
	}
	return targets, nil
}

// Shuffle randomly alters the order of Targets with the provided seed
func (t Targets) Shuffle(seed int64) {
	rand.Seed(seed)
	for i, rnd := range rand.Perm(len(t)) {
		t[i], t[rnd] = t[rnd], t[i]
	}
}
