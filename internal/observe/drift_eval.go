package observe

import "time"

// driftEvalsFile is the name of the JSONL file that stores drift evaluations.
const driftEvalsFile = "drift-evals.jsonl"

// DriftEval records one drift evaluation so detector precision can be measured
// offline. Every scored prompt produces a row carrying the intent it was judged
// against, the keyword sets on both sides, the computed overlap, and whether a
// warning fired — everything a human needs to label the row correct or not.
//
// Evaluations live in their own file rather than observations.jsonl so the
// learning prompts that consume observations see an unchanged stream.
type DriftEval struct {
	Timestamp      time.Time `json:"timestamp"`
	SessionID      string    `json:"session_id"`
	Intent         string    `json:"intent"`
	Prompt         string    `json:"prompt"`
	IntentKeywords []string  `json:"intent_keywords"`
	PromptKeywords []string  `json:"prompt_keywords"`
	Overlap        float64   `json:"overlap"`
	Threshold      float64   `json:"threshold"`
	Edits          int       `json:"edits"`
	Warned         bool      `json:"warned"`
}

// RecordDriftEval appends an evaluation as a JSON line to drift-evals.jsonl.
// It honours the same .disabled marker as Record, so switching observation off
// also stops prompt text reaching disk. Returns nil when disabled.
func (o *Observer) RecordDriftEval(eval DriftEval) error {
	return o.appendJSONL(driftEvalsFile, eval)
}
