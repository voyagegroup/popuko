package main

import (
	"testing"
)

func TestParseCommand1(t *testing.T) {
	ok, cmd := parseCommand("@bot r+")
	if !ok {
		t.Fatal("should be ok")
	}

	v, ok := cmd.(*AcceptChangeByReviewerCommand)
	if !ok {
		t.Fatal("should be AcceptChangeByReviewerCommand")
	}

	if name := v.BotName(); name != "bot" {
		t.Fatalf("should be the expected bot name: %v\n", name)
	}
}

func TestParseCommand2(t *testing.T) {
	ok, cmd := parseCommand("@reviewer r?")
	if !ok {
		t.Fatal("should be ok")
	}

	v, ok := cmd.(*AssignReviewerCommand)
	if !ok {
		t.Fatal("should be AssignReviewerCommand")
	}

	if v.Reviewer != "reviewer" {
		t.Fatal("should be the expected reviewer")
	}
}

func TestParseCommand3(t *testing.T) {
	ok, cmd := parseCommand("@bot r=popuko,pipimi")
	if !ok {
		t.Fatal("should be ok")
	}

	v, ok := cmd.(*AcceptChangeByOthersCommand)
	if !ok {
		t.Fatal("should be AcceptChangeByOthersCommand")
	}

	if name := v.BotName(); name != "bot" {
		t.Fatalf("should be the expected bot name: %v\n", name)
	}

	if name := v.Reviewer[0]; name != "popuko" {
		t.Fatalf("should be the expected reviewer 1: %v\n", name)
	}

	if name := v.Reviewer[1]; name != "pipimi" {
		t.Fatalf("should be the expected reviewer 2: %v\n", name)
	}
}

func TestParseCommand5(t *testing.T) {
	ok, cmd := parseCommand("")
	if ok {
		t.Fatal("should be no result")
	}

	if cmd != nil {
		t.Fatal("command should be nil")
	}
}

func TestParseCommand6(t *testing.T) {
	ok, _ := parseCommand(`@bot
    r+`)
	if ok {
		t.Fatal("should not be ok")
	}
}

func TestParseCommand7(t *testing.T) {
	ok, _ := parseCommand("@bot")
	if ok {
		t.Fatal("should not be ok")
	}
}

func TestParseCommand8(t *testing.T) {
	ok, _ := parseCommand("bot r+")
	if ok {
		t.Fatal("should not be ok")
	}
}

func TestParseCommand9(t *testing.T) {
	ok, _ := parseCommand("Hello, I'm john.")
	if ok {
		t.Fatal("should not be ok")
	}
}

func TestParseCommand10(t *testing.T) {
	ok, _ := parseCommand("    @bot r+")
	if ok {
		t.Fatal("should not be ok")
	}
}

func TestParseCommand11(t *testing.T) {
	ok, _ := parseCommand(`
    @bot r+`)
	if ok {
		t.Fatal("should not be ok")
	}
}
