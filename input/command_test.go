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

	if name := v.Reviewer[0]; name != "popuko" {
		t.Errorf("should be the expected reviewer 1: %v\n", name)
		return
	}

	if name := v.Reviewer[1]; name != "pipimi" {
		t.Errorf("should be the expected reviewer 2: %v\n", name)
		return
	}
}

func TestParseCommand5(t *testing.T) {
	ok, cmd := ParseCommand("")
	if ok {
		t.Errorf("should be no result")
		return
	}

	if cmd != nil {
		t.Errorf("command should be nil")
		return
	}
}

func TestParseCommand6(t *testing.T) {
	ok, _ := ParseCommand(`@bot
    r+`)
	if ok {
		t.Errorf("should not be ok")
		return
	}
}

func TestParseCommand7(t *testing.T) {
	ok, _ := ParseCommand("@bot")
	if ok {
		t.Errorf("should not be ok")
		return
	}
}

func TestParseCommand8(t *testing.T) {
	ok, _ := ParseCommand("bot r+")
	if ok {
		t.Errorf("should not be ok")
		return
	}
}

func TestParseCommand9(t *testing.T) {
	ok, _ := ParseCommand("Hello, I'm john.")
	if ok {
		t.Errorf("should not be ok")
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

func TestParseCommand11(t *testing.T) {
	ok, _ := ParseCommand(`
    @bot r+`)
	if ok {
		t.Errorf("should not be ok")
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
