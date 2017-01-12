package input

import (
	"testing"
)

func TestParseCommand1(t *testing.T) {
	ok, cmd := ParseCommand("@bot r+")
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AcceptChangeByReviewerCommand)
	if !ok {
		t.Errorf("should be AcceptChangeByReviewerCommand")
		return
	}

	if name := v.BotName(); name != "bot" {
		t.Errorf("should be the expected bot name: %v\n", name)
		return
	}
}

func TestParseCommand2(t *testing.T) {
	ok, cmd := ParseCommand("@reviewer r?")
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AssignReviewerCommand)
	if !ok {
		t.Errorf("should be AssignReviewerCommand")
		return
	}

	if v.Reviewer != "reviewer" {
		t.Errorf("should be the expected reviewer")
		return
	}
}

func TestParseCommand3(t *testing.T) {
	ok, cmd := ParseCommand("@bot r=popuko,pipimi")
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AcceptChangeByOthersCommand)
	if !ok {
		t.Errorf("should be AcceptChangeByOthersCommand")
		return
	}

	if name := v.BotName(); name != "bot" {
		t.Errorf("should be the expected bot name: %v\n", name)
		return
	}

	if len(v.Reviewer) == 0 {
		t.Errorf("should have some reviewers")
		return
	}

	if name := v.Reviewer[0]; name != "popuko" {
		t.Errorf("should be the expected reviewer 1: %v\n", name)
		return
	}

	if name := v.Reviewer[1]; name != "pipimi" {
		t.Errorf("should be the expected reviewer 2: %v\n", name)
		return
	}
}

func TestParseCommand10(t *testing.T) {
	ok, cmd := ParseCommand("    @bot r+")
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AcceptChangeByReviewerCommand)
	if !ok {
		t.Errorf("should be AcceptChangeByReviewerCommand")
		return
	}

	if name := v.BotName(); name != "bot" {
		t.Errorf("should be the expected bot name: %v\n", name)
		return
	}
}

func TestParseCommand12(t *testing.T) {
	ok, cmd := ParseCommand("r? @reviewer")
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AssignReviewerCommand)
	if !ok {
		t.Errorf("should be AssignReviewerCommand")
		return
	}

	if v.Reviewer != "reviewer" {
		t.Errorf("should be the expected reviewer")
		return
	}
}

func TestParseCommand13(t *testing.T) {
	ok, cmd := ParseCommand("@bot        r+")
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AcceptChangeByReviewerCommand)
	if !ok {
		t.Errorf("should be AcceptChangeByReviewerCommand")
		return
	}

	if name := v.BotName(); name != "bot" {
		t.Errorf("should be the expected bot name: %v\n", name)
		return
	}
}

func TestParseCommand14(t *testing.T) {
	ok, cmd := ParseCommand(`@bot        r+



	`)
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AcceptChangeByReviewerCommand)
	if !ok {
		t.Errorf("should be AcceptChangeByReviewerCommand")
		return
	}

	if name := v.BotName(); name != "bot" {
		t.Errorf("should be the expected bot name: %v\n", name)
		return
	}
}

func TestParseCommand15(t *testing.T) {
	ok, cmd := ParseCommand("@bot　　　  r+")
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AcceptChangeByReviewerCommand)
	if !ok {
		t.Errorf("should be AcceptChangeByReviewerCommand")
		return
	}

	if name := v.BotName(); name != "bot" {
		t.Errorf("should be the expected bot name: %v\n", name)
		return
	}
}

func TestParseCommand16(t *testing.T) {
	ok, cmd := ParseCommand("@bot r-")
	v, ok := cmd.(*CancelApprovedByReviewerCommand)
	if !ok {
		t.Errorf("should be CancelApprovedByReviewerCommand")
		return
	}

	if name := v.BotName(); name != "bot" {
		t.Errorf("should be the expected bot name: %v\n", name)
		return
	}
}

func TestParseCommand17(t *testing.T) {
	ok, cmd := ParseCommand("@bot-bot r-")
	if !ok {
		t.Errorf("should be success to parse")
		return
	}

	v, ok := cmd.(*CancelApprovedByReviewerCommand)
	if !ok {
		t.Errorf("should be CancelApprovedByReviewerCommand")
		return
	}

	if name := v.BotName(); name != "bot-bot" {
		t.Errorf("should be the expected bot name: %v\n", name)
		return
	}
}

func TestParseCommand18(t *testing.T) {
	ok, cmd := ParseCommand("@bot-bot r+")
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AcceptChangeByReviewerCommand)
	if !ok {
		t.Errorf("should be AcceptChangeByReviewerCommand")
		return
	}

	if name := v.BotName(); name != "bot-bot" {
		t.Errorf("should be the expected bot name: %v\n", name)
		return
	}
}

func TestParseCommand19(t *testing.T) {
	ok, cmd := ParseCommand("r? @reviewer-a")
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AssignReviewerCommand)
	if !ok {
		t.Errorf("should be AssignReviewerCommand")
		return
	}

	if v.Reviewer != "reviewer-a" {
		t.Errorf("should be the expected reviewer")
		return
	}
}

func TestParseCommand20(t *testing.T) {
	ok, cmd := ParseCommand("@reviewer-a r?")
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AssignReviewerCommand)
	if !ok {
		t.Errorf("should be AssignReviewerCommand")
		return
	}

	if v.Reviewer != "reviewer-a" {
		t.Errorf("should be the expected reviewer")
		return
	}
}

func TestParseCommand21(t *testing.T) {
	ok, cmd := ParseCommand("@bot r=popuko-a,pipimi-b")
	if !ok {
		t.Errorf("should be ok")
		return
	}

	v, ok := cmd.(*AcceptChangeByOthersCommand)
	if !ok {
		t.Errorf("should be AcceptChangeByOthersCommand")
		return
	}

	if name := v.BotName(); name != "bot" {
		t.Errorf("should be the expected bot name: %v\n", name)
		return
	}

	if name := v.Reviewer[0]; name != "popuko-a" {
		t.Errorf("should be the expected reviewer 1: %v\n", name)
		return
	}

	if name := v.Reviewer[1]; name != "pipimi-b" {
		t.Errorf("should be the expected reviewer 2: %v\n", name)
		return
	}
}

func TestParseCommandInvalidCase(t *testing.T) {
	input := []string{
		"Hello, I'm john.",
		"",
		"bot r+",

		"@bot",

		" @ bot r+",
		" @ bot r +",
		`
    @bot r+`,
		`@bot
    r+`,

		" @ bot r-",
		" @ bot r -",
		`
    @bot r-`,
		`@bot
    r-`,

		" @ bot r=a",
		" @ bot r = a",
		" @ bot r =a",
		`
    @bot r=a`,
		`@bot
    r=a`,
	}
	for _, item := range input {
		if ok, _ := ParseCommand(item); ok {
			t.Errorf("%v should not be ok", item)
		}
	}
}
