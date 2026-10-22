package config

import (
	"fmt"
	"strings"
	"testing"
)

// A name the Router cannot block would be ignored at dispatch, so the
// configuration is refused instead of carrying a policy that cannot apply.
func TestRequestParamsRefusesABlockedParamTheRouterCannotBlock(t *testing.T) {
	template := `routing:
  decisions:
    - name: bounded
      plugins:
        - type: request_params
          configuration:
            blocked_params:
              - %s
`
	for _, name := range []string{"logprobs", "top_logprobs", "max_token", "model", "messages"} {
		if _, err := ParseRoutingYAMLBytes([]byte(fmt.Sprintf(template, name))); err == nil {
			t.Errorf("accepted a blocked param the Router cannot block: %s", name)
		} else if !strings.Contains(err.Error(), name) {
			t.Errorf("the error for %s does not name it: %v", name, err)
		}
	}
	for _, name := range []string{"verbosity", "text.verbosity", "stream_options", "user", "temperature", "seed"} {
		if _, err := ParseRoutingYAMLBytes([]byte(fmt.Sprintf(template, name))); err != nil {
			t.Errorf("rejected a blockable param %s: %v", name, err)
		}
	}
}
