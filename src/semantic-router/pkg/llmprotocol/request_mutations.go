package llmprotocol

import (
	"fmt"
	"strings"
)

// DefaultOutputTokens supplies policy only when the caller has no effective
// output bound. Model capacity metadata is never a provider default.
func DefaultOutputTokens(request *Request, value *int) bool {
	if request == nil || request.Sampling.MaxOutputTokens != nil || value == nil || *value <= 0 {
		return false
	}
	request.Sampling.MaxOutputTokens = Int64(int64(*value))
	return true
}

// DefaultAutomaticOutput preserves caller limits and defers a missing bound
// until the provider has rendered the complete model-specific input.
func DefaultAutomaticOutput(request *Request) bool {
	if request == nil || request.Sampling.MaxOutputTokens != nil || request.Sampling.AutomaticOutput {
		return false
	}
	request.Sampling.AutomaticOutput = true
	return true
}

// CapOutputTokens clamps an explicit request value; it never invents a limit.
func CapOutputTokens(request *Request, limit *int) bool {
	if request != nil && request.Sampling.AutomaticOutput && limit != nil {
		request.Sampling.AutomaticOutputCap = Int64(int64(*limit))
	}
	if limit == nil || request.Sampling.MaxOutputTokens == nil || *request.Sampling.MaxOutputTokens <= int64(*limit) {
		return false
	}
	request.Sampling.MaxOutputTokens = Int64(int64(*limit))
	return true
}

func CapCandidateCount(request *Request, limit *int) bool {
	if limit == nil || request.CandidateCount == nil || *request.CandidateCount <= int64(*limit) {
		return false
	}
	request.CandidateCount = Int64(int64(*limit))
	return true
}

// requestBlockers maps every request field a policy may block to the removal
// of that field. The keys are the names the client sends, so the map is the
// whole vocabulary a `blocked_params` entry may use.
var requestBlockers = map[string]func(*Request) bool{
	"frequency_penalty": func(r *Request) bool { return clearFloat(&r.Sampling.FrequencyPenalty) },
	"presence_penalty":  func(r *Request) bool { return clearFloat(&r.Sampling.PresencePenalty) },
	"max_tokens":        func(r *Request) bool { return clearInt(&r.Sampling.MaxOutputTokens) },
	"max_completion_tokens": func(r *Request) bool {
		return clearInt(&r.Sampling.MaxOutputTokens)
	},
	"max_output_tokens": func(r *Request) bool { return clearInt(&r.Sampling.MaxOutputTokens) },
	"n":                 func(r *Request) bool { return clearInt(&r.CandidateCount) },
	"candidate_count":   func(r *Request) bool { return clearInt(&r.CandidateCount) },
	"response_format":   func(r *Request) bool { return clearOutputFormat(&r.OutputFormat) },
	"output_format":     func(r *Request) bool { return clearOutputFormat(&r.OutputFormat) },
	"seed":              func(r *Request) bool { return clearInt(&r.Sampling.Seed) },
	"stop":              func(r *Request) bool { return clearStop(&r.Sampling.Stop) },
	"temperature":       func(r *Request) bool { return clearFloat(&r.Sampling.Temperature) },
	"top_p":             func(r *Request) bool { return clearFloat(&r.Sampling.TopP) },
	"min_p":             func(r *Request) bool { return clearFloat(&r.Sampling.MinP) },
	"repetition_penalty": func(r *Request) bool {
		return clearFloat(&r.Sampling.RepetitionPenalty)
	},
	"cache_salt":       func(r *Request) bool { return clearStringPointer(&r.CacheSalt) },
	"prompt_cache_key": func(r *Request) bool { return clearString(&r.PromptCacheKey) },
	"top_k":            func(r *Request) bool { return clearInt(&r.Sampling.TopK) },
	"tools":            func(r *Request) bool { return clearTools(&r.Tools) },
	"tool_choice":      func(r *Request) bool { return clearToolChoice(&r.ToolChoice) },
	"parallel_tool_calls": func(r *Request) bool {
		return clearBool(&r.ParallelToolCalls)
	},
	"reasoning_effort":        func(r *Request) bool { return clearString(&r.ReasoningEffort) },
	"reasoning_budget_tokens": func(r *Request) bool { return clearInt(&r.ReasoningBudgetTokens) },
	"metadata":                func(r *Request) bool { return clearMetadata(&r.Metadata) },
	"store":                   func(r *Request) bool { return clearBool(&r.Store) },
	"stream":                  func(r *Request) bool { return clearStream(&r.Stream) },
	"verbosity":               func(r *Request) bool { return clearString(&r.TextVerbosity) },
	"text.verbosity":          func(r *Request) bool { return clearString(&r.TextVerbosity) },
	"text_verbosity":          func(r *Request) bool { return clearString(&r.TextVerbosity) },
	"stream_options": func(r *Request) bool {
		changed := r.StreamOptions.IncludeUsage != nil || r.StreamOptions.IncludeObfuscation != nil
		r.StreamOptions = StreamOptions{}
		return changed
	},
	"user":       func(r *Request) bool { return clearString(&r.EndUserID) },
	"truncation": func(r *Request) bool { return clearString(&r.Truncation) },
	"previous_response_id": func(r *Request) bool {
		return clearString(&r.PreviousResponseID)
	},
	"conversation":      func(r *Request) bool { return clearString(&r.ConversationID) },
	"conversation_id":   func(r *Request) bool { return clearString(&r.ConversationID) },
	"auto_store":        func(r *Request) bool { return clearBool(&r.AutoStore) },
	"reasoning.summary": func(r *Request) bool { return clearString(&r.ReasoningSummary) },
	"reasoning_summary": func(r *Request) bool { return clearString(&r.ReasoningSummary) },
	"reasoning.display": func(r *Request) bool { return clearString(&r.ReasoningDisplay) },
	"reasoning_display": func(r *Request) bool { return clearString(&r.ReasoningDisplay) },
}

func clearFloat(value **float64) bool {
	changed := *value != nil
	*value = nil
	return changed
}

func clearInt(value **int64) bool {
	changed := *value != nil
	*value = nil
	return changed
}

func clearBool(value **bool) bool {
	changed := *value != nil
	*value = nil
	return changed
}

func clearString(value *string) bool {
	changed := *value != ""
	*value = ""
	return changed
}

func clearStringPointer(value **string) bool {
	changed := *value != nil
	*value = nil
	return changed
}

func clearStop(value *[]string) bool {
	changed := len(*value) > 0
	*value = nil
	return changed
}

func clearTools(value *[]Tool) bool {
	changed := len(*value) > 0
	*value = nil
	return changed
}

func clearMetadata(value *map[string]string) bool {
	changed := len(*value) > 0
	*value = nil
	return changed
}

func clearOutputFormat(value *OutputFormat) bool {
	changed := value.Kind != ""
	*value = OutputFormat{}
	return changed
}

func clearToolChoice(value *ToolChoice) bool {
	changed := value.Mode != "" || value.Name != ""
	*value = ToolChoice{}
	return changed
}

func clearStream(value *bool) bool {
	changed := *value
	*value = false
	return changed
}

// BlockRequestField removes a named optional field from the neutral request.
//
// A name the Router does not model names no field, so it removes nothing: a
// policy that names one is refused when the configuration loads, not here.
func BlockRequestField(request *Request, field string) (bool, error) {
	switch field {
	case "":
		return false, nil
	case "model", "messages":
		return false, fmt.Errorf("required semantic field %q cannot be blocked", field)
	}
	blocker, ok := requestBlockers[field]
	if !ok {
		return false, nil
	}
	return blocker(request), nil
}

// BlockableRequestField reports whether a policy may block the named field.
// It answers with the same vocabulary BlockRequestField implements, so the
// configuration validation and the plugin agree on what a name means.
func BlockableRequestField(field string) bool {
	switch strings.TrimSpace(field) {
	case "", "model", "messages":
		return false
	}
	_, ok := requestBlockers[strings.TrimSpace(field)]
	return ok
}
