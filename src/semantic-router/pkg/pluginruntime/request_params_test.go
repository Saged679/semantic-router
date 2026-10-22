package pluginruntime

import (
	"testing"

	"github.com/vllm-project/semantic-router/src/semantic-router/pkg/config"
	"github.com/vllm-project/semantic-router/src/semantic-router/pkg/llmprotocol"
)

func TestRequestParamsBlockedOutputIsDefaultedBeforeCap(t *testing.T) {
	defaultTokens, maxTokens, maxN := 100, 50, 1
	request := &llmprotocol.Request{CandidateCount: llmprotocol.Int64(3), Sampling: llmprotocol.Sampling{MaxOutputTokens: llmprotocol.Int64(1000)}}
	result, err := ApplyRequestParams(request, &config.RequestParamsPluginConfig{BlockedParams: []string{"max_tokens"}, DefaultMaxTokens: config.FixedOutputTokenDefault(defaultTokens), MaxTokensLimit: &maxTokens, MaxN: &maxN})
	if err != nil {
		t.Fatal(err)
	}
	if *request.Sampling.MaxOutputTokens != 50 || *request.CandidateCount != 1 || !result.DefaultedOutputTokens || !result.CappedOutputTokens || !result.CappedCandidateCount {
		t.Fatalf("wrong policy effects: %+v, request %+v", result, request)
	}
}

func TestRequestParamsBlocksPromptCacheKey(t *testing.T) {
	request := &llmprotocol.Request{PromptCacheKey: "codex-session"}
	result, err := ApplyRequestParams(request, &config.RequestParamsPluginConfig{BlockedParams: []string{"prompt_cache_key"}})
	if err != nil {
		t.Fatal(err)
	}
	if request.PromptCacheKey != "" || !result.Changed || len(result.Blocked) != 1 || result.Blocked[0] != "prompt_cache_key" {
		t.Fatalf("request params did not block prompt_cache_key: request=%+v result=%+v", request, result)
	}
}

// A field the Router forwards is the one a policy names, so the parameter
// never reaches the backend.
func TestRequestParamsBlocksTheForwardedFields(t *testing.T) {
	usage := true
	verbosity := "high"
	request := &llmprotocol.Request{
		TextVerbosity: verbosity,
		StreamOptions: llmprotocol.StreamOptions{IncludeUsage: &usage},
		EndUserID:     "user-1",
	}
	result, err := ApplyRequestParams(request, &config.RequestParamsPluginConfig{
		BlockedParams: []string{"verbosity", "stream_options", "user"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if request.TextVerbosity != "" || request.StreamOptions.IncludeUsage != nil || request.EndUserID != "" {
		t.Fatalf("request params did not block the forwarded fields: request=%+v", request)
	}
	if !result.Changed || len(result.Blocked) != 3 {
		t.Fatalf("request params did not report the blocked fields: %+v", result)
	}
}
