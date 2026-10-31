package spec

import (
	"strings"
	"testing"
)

// Hints of R18-R22 that many cases share.
const (
	hintShowsExist = `name only requirements that exist, e.g. "Shows: R1, R2"`
	hintShowsList  = `write it as e.g. "Shows: R1, R2"`
	hintKind       = "use unit, api, e2e, or manual"
	hintDecOrder   = "put them in this order: Why:, Not chosen:, Cost:, Rule:, Replaces:"
)

// removedR2 adds a removed R2 after R1.
var removedR2 = change{"  Check: e2e\n", "  Check: e2e\n- **R2.** (removed in R001: merged into R1)\n"}

// TestCheckShows checks each message of R18. S1's Shows: line is line 20.
func TestCheckShows(t *testing.T) {
	shows := "  Shows: R1\n"
	notList := Problem{20, "S1's Shows: line does not list requirements", hintShowsList}
	runCases(t, []checkCase{
		{"no Shows: line",
			[]change{{shows, ""}},
			[]Problem{{18, "S1 has no Shows: line", `add a line "Shows: R1" under S1, naming the requirements it shows`}}},
		{"two Shows: lines",
			[]change{{shows, shows + shows}},
			[]Problem{{21, "S1 has two Shows: lines", "merge them into one"}}},
		{"two Shows: lines: only the first is checked",
			[]change{{shows, shows + "  Shows: R9\n"}},
			[]Problem{{21, "S1 has two Shows: lines", "merge them into one"}}},
		{"not split by commas",
			[]change{{"Shows: R1", "Shows: R1 R1"}},
			[]Problem{notList}},
		{"empty part",
			[]change{{"Shows: R1", "Shows: R1,,R1"}},
			[]Problem{notList}},
		{"trailing comma",
			[]change{{"Shows: R1", "Shows: R1,"}},
			[]Problem{notList}},
		{"no value",
			[]change{{"Shows: R1", "Shows:"}},
			[]Problem{notList}},
		{"lower case id",
			[]change{{"Shows: R1", "Shows: r1"}},
			[]Problem{notList}},
		{"leading zero",
			[]change{{"Shows: R1", "Shows: R01"}},
			[]Problem{notList}},
		{"not a list: its ids are not looked up",
			[]change{{"Shows: R1", "Shows: R9, all"}},
			[]Problem{notList}},
		{"an id twice counts as removed only when every copy is removed",
			[]change{
				{"Shows: R1", "Shows: R1, R2"},
				{"  Check: e2e\n", "  Check: e2e\n- **R2.** (removed in R001: x)\n- **R2.** y.\n  Check: e2e\n"},
			},
			[]Problem{{27, "R2 appears twice (first on line 26)", hintNumbers}}},
		{"no such requirement",
			[]change{{"Shows: R1", "Shows: R1, R9"}},
			[]Problem{{20, "S1 shows R9, but there is no R9", hintShowsExist}}},
		{"one problem per missing id, each once",
			[]change{{"Shows: R1", "Shows: R9, R1, R8, R9"}},
			[]Problem{
				{20, "S1 shows R9, but there is no R9", hintShowsExist},
				{20, "S1 shows R8, but there is no R8", hintShowsExist},
			}},
		{"spaces around ids",
			[]change{{"Shows: R1", "Shows:   R1 ,R1  "}},
			nil},
		{"removed requirement",
			[]change{removedR2, {"Shows: R1", "Shows: R1, R2"}},
			[]Problem{{20, "S1 shows R2, but R2 is removed", "remove R2 from the Shows: line"}}},
		{"ids on a continuation line",
			[]change{{"Shows: R1", "Shows: R1,\n    R9"}},
			[]Problem{{20, "S1 shows R9, but there is no R9", hintShowsExist}}},
		{"requirement in another section still exists",
			[]change{
				{"## Decisions\n\nNone.\n", "## Decisions\n\n- **R2.** sliceflow does more.\n  Check: e2e\n"},
				{"Shows: R1", "Shows: R1, R2"},
			},
			[]Problem{{33, "R2 is in the wrong section", `move it to "## Requirements"`}}},
		{"placeholder: only R14",
			[]change{{"Shows: R1", "Shows: {{R3}}"}},
			[]Problem{{20, msgPlace + "{{R3}}", hintPlace}}},
		{"placeholder on a continuation line: only R14",
			[]change{{"Shows: R1", "Shows: R1,\n    {{R3}}"}},
			[]Problem{{21, msgPlace + "{{R3}}", hintPlace}}},
		{"{{ in inline code is content, not a placeholder",
			[]change{{"Shows: R1", "Shows: `{{R1}}`"}},
			[]Problem{notList}},
		{"removed scenario has no labels",
			[]change{{shows, shows + "- **S2.** (removed in R001: merged into S1)\n"}},
			nil},
	})
}

// TestCheckCheckKinds checks each message of R19. R1's Check: line is
// line 25.
func TestCheckCheckKinds(t *testing.T) {
	check := "  Check: e2e\n"
	unknown := func(kind string) Problem {
		return Problem{25, `R1 has an unknown check kind "` + kind + `"`, hintKind}
	}
	runCases(t, []checkCase{
		{"no Check: line (S2)",
			[]change{{check, ""}},
			[]Problem{{24, "R1 has no Check: line", `add a line "Check: e2e" under R1. The kinds are unit, api, e2e, and manual`}}},
		{"two Check: lines",
			[]change{{check, check + check}},
			[]Problem{{26, "R1 has two Check: lines", "merge them into one"}}},
		{"two Check: lines: only the first is checked",
			[]change{{check, check + "  Check: tests\n"}},
			[]Problem{{26, "R1 has two Check: lines", "merge them into one"}}},
		{"every kind",
			[]change{{"Check: e2e", "Check: unit, api,e2e ,  manual"}},
			nil},
		{"unknown kinds, each once, in order",
			[]change{{"Check: e2e", "Check: unit, E2E, tests, E2E"}},
			[]Problem{unknown("E2E"), unknown("tests")}},
		{"no value",
			[]change{{"Check: e2e", "Check:"}},
			[]Problem{unknown("")}},
		{"trailing comma",
			[]change{{"Check: e2e", "Check: e2e,"}},
			[]Problem{unknown("")}},
		{"kinds on a continuation line",
			[]change{{"Check: e2e", "Check: unit,\n    e2e, foo"}},
			[]Problem{unknown("foo")}},
		{"placeholder: only R14",
			[]change{{"Check: e2e", "Check: {{unit, api, e2e, or manual}}"}},
			[]Problem{{25, msgPlace + "{{unit, api, e2e, or manual}}", hintPlace}}},
		{"removed requirement has no labels",
			[]change{removedR2},
			nil},
	})
}

// decision is a valid SD1 in "## Decisions". Its lines: SD1 33, Why 34,
// Not chosen 35, Cost 36, Rule 37, Replaces 38.
var decision = change{"## Decisions\n\nNone.\n", "## Decisions\n\n- **SD1.** Demo is small.\n" +
	"  Why: small is easy\n" +
	"  Not chosen: big, because it is slow\n" +
	"  Cost: None\n" +
	"  Rule: R1\n" +
	"  Replaces: device-pairing/SD3\n"}

// TestCheckDecisions checks each message of R20.
func TestCheckDecisions(t *testing.T) {
	why := "  Why: small is easy\n"
	notChosen := "  Not chosen: big, because it is slow\n"
	cost := "  Cost: None\n"
	rule := "  Rule: R1\n"
	replaces := "  Replaces: device-pairing/SD3\n"
	outOfOrder := Problem{33, "SD1's lines are out of order", hintDecOrder}
	runCases(t, []checkCase{
		{"valid decision",
			[]change{decision},
			nil},
		{"no Replaces: is fine",
			[]change{decision, {replaces, ""}},
			nil},
		{"value over two lines",
			[]change{decision, {why, "  Why: small\n    is easy\n"}},
			nil},
		{"no Why: line",
			[]change{decision, {why, ""}},
			[]Problem{{33, "SD1 has no Why: line", `add a line "Why: ..." under SD1`}}},
		{"no Not chosen: line",
			[]change{decision, {notChosen, ""}},
			[]Problem{{33, "SD1 has no Not chosen: line", `add a line "Not chosen: ..." under SD1`}}},
		{"no Cost: line has its own hint",
			[]change{decision, {cost, ""}},
			[]Problem{{33, "SD1 has no Cost: line", `add a line "Cost: ..." under SD1: the downside we accept, or None`}}},
		{"no Rule: line",
			[]change{decision, {rule, ""}},
			[]Problem{{33, "SD1 has no Rule: line", `add a line "Rule: ..." under SD1`}}},
		{"two Why: lines",
			[]change{decision, {why, why + "  Why: again\n"}},
			[]Problem{{35, "SD1 has two Why: lines", "keep one"}}},
		{"two Why: lines, the second out of place: no out of order",
			[]change{decision, {rule, rule + "  Why: again\n"}},
			[]Problem{{38, "SD1 has two Why: lines", "keep one"}}},
		{"two Replaces: lines",
			[]change{decision, {replaces, replaces + replaces}},
			[]Problem{{39, "SD1 has two Replaces: lines", "keep one"}}},
		{"empty Why:",
			[]change{decision, {why, "  Why:\n"}},
			[]Problem{{34, "SD1's Why: line is empty", "write its value"}}},
		{"empty Cost:",
			[]change{decision, {cost, "  Cost:   \n"}},
			[]Problem{{36, "SD1's Cost: line is empty", "write its value"}}},
		{"empty Rule: gives no R21 problem",
			[]change{decision, {rule, "  Rule:\n"}},
			[]Problem{{37, "SD1's Rule: line is empty", "write its value"}}},
		{"empty Replaces: gives no R21 problem",
			[]change{decision, {replaces, "  Replaces:\n"}},
			[]Problem{{38, "SD1's Replaces: line is empty", "write its value"}}},
		{"empty second line is not checked",
			[]change{decision, {why, why + "  Why:\n"}},
			[]Problem{{35, "SD1 has two Why: lines", "keep one"}}},
		{"Why: and Not chosen: swapped",
			[]change{decision, {why + notChosen, notChosen + why}},
			[]Problem{outOfOrder}},
		{"Replaces: not last",
			[]change{decision, {rule + replaces, replaces + rule}},
			[]Problem{outOfOrder}},
		{"missing label and out of order: two problems",
			[]change{decision, {why + notChosen + cost, cost + notChosen}},
			[]Problem{
				{33, "SD1 has no Why: line", `add a line "Why: ..." under SD1`},
				outOfOrder,
			}},
		{"other labels are not checked",
			[]change{decision, {rule, rule + "  Check: e2e\n"}},
			nil},
		{"removed decision has no labels",
			[]change{{"## Decisions\n\nNone.\n", "## Decisions\n\n- **SD1.** (removed in R001: no longer needed)\n"}},
			nil},
	})
}

// TestCheckRules checks each message of R21. Rule: is line 37, Replaces:
// line 38.
func TestCheckRules(t *testing.T) {
	notRule := Problem{37, `SD1's Rule: is not "none" or a list of requirements`, `write it as e.g. "Rule: R11", or "Rule: none"`}
	notReplaces := Problem{38, `SD1's Replaces: is not "<slice>/SD<n>"`, `write it as e.g. "Replaces: device-pairing/SD3"`}
	rule := func(value string) change { return change{"  Rule: R1\n", "  Rule: " + value + "\n"} }
	replaces := func(value string) change {
		return change{"  Replaces: device-pairing/SD3\n", "  Replaces: " + value + "\n"}
	}
	runCases(t, []checkCase{
		{"Rule: none",
			[]change{decision, rule("none")},
			nil},
		{"Rule: a list",
			[]change{decision, {"  Check: e2e\n", "  Check: e2e\n- **R2.** sliceflow stops.\n  Check: e2e\n"}, rule("R1, R2")},
			nil},
		{"Rule: None is not none",
			[]change{decision, rule("None")},
			[]Problem{notRule}},
		{"Rule: not split by commas",
			[]change{decision, rule("R1 R1")},
			[]Problem{notRule}},
		{"Rule: none in a list",
			[]change{decision, rule("none, R1")},
			[]Problem{notRule}},
		{"Rule: trailing comma",
			[]change{decision, rule("R1,")},
			[]Problem{notRule}},
		{"Rule: no such requirement",
			[]change{decision, rule("R1, R9")},
			[]Problem{{37, "SD1's Rule: names R9, but there is no R9", hintRuleExist}}},
		{"Rule: removed requirement (the removed R2 moves Rule: to line 38)",
			[]change{decision, removedR2, rule("R2")},
			[]Problem{{38, "SD1's Rule: names R2, but R2 is removed", hintRuleExist}}},
		{"Rule: ids on a continuation line",
			[]change{decision, rule("R1,\n    R9")},
			[]Problem{{37, "SD1's Rule: names R9, but there is no R9", hintRuleExist}}},
		{"Rule: placeholder: only R14",
			[]change{decision, rule("{{R11}}")},
			[]Problem{{37, msgPlace + "{{R11}}", hintPlace}}},
		{"Replaces: digits and dashes",
			[]change{decision, replaces("0002-spec-and-state/SD12")},
			nil},
		{"Replaces: no /SD<n>",
			[]change{decision, replaces("device-pairing")},
			[]Problem{notReplaces}},
		{"Replaces: lower case sd",
			[]change{decision, replaces("device-pairing/sd3")},
			[]Problem{notReplaces}},
		{"Replaces: SD0",
			[]change{decision, replaces("device-pairing/SD0")},
			[]Problem{notReplaces}},
		{"Replaces: upper case slice",
			[]change{decision, replaces("Device-Pairing/SD3")},
			[]Problem{notReplaces}},
		{"Replaces: slice starts with a dash",
			[]change{decision, replaces("-pairing/SD3")},
			[]Problem{notReplaces}},
		{"Replaces: slice name too long",
			[]change{decision, replaces(strings.Repeat("b", 65) + "/SD3")},
			[]Problem{notReplaces}},
		{"Replaces: slice name of 64 characters",
			[]change{decision, replaces(strings.Repeat("b", 64) + "/SD3")},
			nil},
		{"Replaces: placeholder: only R14",
			[]change{decision, replaces("{{x}}")},
			[]Problem{{38, msgPlace + "{{x}}", hintPlace}}},
		{"removed decision",
			[]change{{"## Decisions\n\nNone.\n", "## Decisions\n\n- **SD1.** (removed in R001: no longer needed)\n"}},
			nil},
	})
}

// TestCheckQuestions checks the message of R22.
func TestCheckQuestions(t *testing.T) {
	questions := "## Open questions\n\nNone.\n"
	runCases(t, []checkCase{
		{"open question (S3)",
			[]change{{questions, "## Open questions\n\n- **Q1.** Is it done?\n"}},
			[]Problem{{37, "Q1 is not answered", "write the answer into the spec, and remove Q1. When no question is left, write None"}}},
		{"two open questions",
			[]change{{questions, "## Open questions\n\n- **Q1.** Is it done?\n- **Q2.** Is it fast?\n"}},
			[]Problem{
				{37, "Q1 is not answered", "write the answer into the spec, and remove Q1. When no question is left, write None"},
				{38, "Q2 is not answered", "write the answer into the spec, and remove Q2. When no question is left, write None"},
			}},
		{"removed question",
			[]change{{questions, "## Open questions\n\n- **Q1.** (removed in R001: answered in R1)\n"}},
			nil},
	})
}

// TestCheckItemsOrder checks the order of problems on one line, and S2's
// two messages (R2).
func TestCheckItemsOrder(t *testing.T) {
	runCases(t, []checkCase{
		{"S2: no Check: line, and Shows: names a missing requirement",
			[]change{{"  Check: e2e\n", ""}, {"Shows: R1", "Shows: R9"}},
			[]Problem{
				{20, "S1 shows R9, but there is no R9", hintShowsExist},
				{24, "R1 has no Check: line", `add a line "Check: e2e" under R1. The kinds are unit, api, e2e, and manual`},
			}},
		{"R15 before R18 on a scenario",
			[]change{{"- **S1.**", "- **S2.**"}, {"  Shows: R1\n", ""}},
			[]Problem{
				{18, "S2 comes first: expected S1", hintNum},
				{18, "S2 has no Shows: line", `add a line "Shows: R1" under S2, naming the requirements it shows`},
			}},
		{"R16 before R22 on a question",
			[]change{{"## Out of scope\n\nNone.\n", "## Out of scope\n\n- **Q1.** Is it done?\n"}},
			[]Problem{
				{29, "Q1 is in the wrong section", `move it to "## Open questions"`},
				{29, "Q1 is not answered", "write the answer into the spec, and remove Q1. When no question is left, write None"},
			}},
		{"R20 missing label before out of order, both on the decision",
			[]change{decision, {"  Why: small is easy\n  Not chosen: big, because it is slow\n", "  Not chosen: big, because it is slow\n"},
				{"  Cost: None\n  Rule: R1\n", "  Rule: R1\n  Cost: None\n"}},
			[]Problem{
				{33, "SD1 has no Why: line", `add a line "Why: ..." under SD1`},
				{33, "SD1's lines are out of order", hintDecOrder},
			}},
	})
}
