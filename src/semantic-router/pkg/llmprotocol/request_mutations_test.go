package llmprotocol

import "testing"

// A policy names the fields it removes in the names the client sends. Every
// name it accepts must remove the field it names, and a name that names
// nothing must be refused rather than ignored.
func TestBlockRequestFieldRemovesTheFieldItNames(t *testing.T) {
	usage := true
	seed := int64(7)
	verbosity := "high"
	temperature := 0.7
	request := &Request{
		Model:         "m",
		TextVerbosity: verbosity,
		StreamOptions: StreamOptions{IncludeUsage: &usage},
		EndUserID:     "u",
		Truncation:    "auto",
		Sampling:      Sampling{Seed: &seed, Temperature: &temperature},
	}

	blocked, err := BlockRequestField(request, "verbosity")
	if err != nil {
		t.Fatal(err)
	}
	if !blocked || request.TextVerbosity != "" {
		t.Errorf("verbosity was not removed: %q", request.TextVerbosity)
	}
	blocked, err = BlockRequestField(request, "stream_options")
	if err != nil {
		t.Fatal(err)
	}
	if !blocked || request.StreamOptions.IncludeUsage != nil {
		t.Errorf("stream_options was not removed: %+v", request.StreamOptions)
	}
	blocked, err = BlockRequestField(request, "user")
	if err != nil {
		t.Fatal(err)
	}
	if !blocked || request.EndUserID != "" {
		t.Errorf("user was not removed: %q", request.EndUserID)
	}
	blocked, err = BlockRequestField(request, "truncation")
	if err != nil {
		t.Fatal(err)
	}
	if !blocked || request.Truncation != "" {
		t.Errorf("truncation was not removed: %q", request.Truncation)
	}
	blocked, err = BlockRequestField(request, "seed")
	if err != nil {
		t.Fatal(err)
	}
	if !blocked || request.Sampling.Seed != nil {
		t.Errorf("seed was not removed: %v", request.Sampling.Seed)
	}
	// A field the request does not carry changes nothing.
	blocked, err = BlockRequestField(request, "top_k")
	if err != nil {
		t.Fatal(err)
	}
	if blocked {
		t.Error("a field the request does not carry reported a change")
	}
}

// A name the Router does not model names no field to remove, so a policy
// that names one removes nothing. The configuration validation refuses such a
// name when the configuration loads.
func TestBlockRequestFieldIgnoresANameItCannotBlock(t *testing.T) {
	for _, name := range []string{"logprobs", "top_logprobs", "max_token", "unknown_field"} {
		blocked, err := BlockRequestField(&Request{}, name)
		if err != nil {
			t.Errorf("a name the Router does not model was refused: %s: %v", name, err)
		}
		if blocked {
			t.Errorf("a name the Router does not model reported a change: %s", name)
		}
	}
	for _, name := range []string{"model", "messages"} {
		if _, err := BlockRequestField(&Request{}, name); err == nil {
			t.Errorf("accepted a required field: %s", name)
		}
	}
}

// The configuration validation and the plugin read the same vocabulary.
func TestBlockableRequestFieldAnswersLikeThePlugin(t *testing.T) {
	for _, name := range []string{"verbosity", "text.verbosity", "stream_options", "user", "temperature"} {
		if !BlockableRequestField(name) {
			t.Errorf("a field the plugin blocks was refused: %s", name)
		}
	}
	for _, name := range []string{"logprobs", "max_token", "model", "messages"} {
		if BlockableRequestField(name) {
			t.Errorf("a field the plugin cannot block was accepted: %s", name)
		}
	}
	// An empty name blocks nothing, so it is no name at all.
	if BlockableRequestField("") {
		t.Error("an empty name was accepted")
	}
}
