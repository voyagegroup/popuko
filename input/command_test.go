package input

import (
	"testing"
)

func TestParseCommandValidCaseForAcceptChangeByReviewerCommand(t *testing.T) {
	type TestCase struct {
		input           string
		expectedBotName string
	}

	list := []TestCase{
		TestCase{
			input:           "@bot r+",
			expectedBotName: "bot",
		},
		TestCase{
			input:           "@bot-bot r+",
			expectedBotName: "bot-bot",
		},

		TestCase{
			input:           "    @bot r+",
			expectedBotName: "bot",
		},

		TestCase{
			input:           "@bot        r+",
			expectedBotName: "bot",
		},

		TestCase{
			input: `@bot        r+



	`,
			expectedBotName: "bot",
		},
	}
	for _, testcase := range list {
		input := testcase.input

		ok, cmd := ParseCommand(input)
		if !ok {
			t.Errorf("input: `%v` should be ok", input)
			continue
		}

		v, ok := cmd.(*AcceptChangeByReviewerCommand)
		if !ok {
			t.Errorf("input: `%v` should be AcceptChangeByReviewerCommand", input)
			continue
		}

		expected := testcase.expectedBotName
		if actual := v.BotName(); actual != expected {
			t.Errorf("input: `%v` should be the expected bot (`%v`) name but `%v`", input, expected, actual)
			continue
		}
	}
}

func TestParseCommandValidCaseForAcceptChangeByOthersCommand(t *testing.T) {
	type TestCase struct {
		input    string
		expected []string
	}

	list := []TestCase{
		TestCase{
			input:    "@bot r=popuko",
			expected: []string{"popuko"},
		},
		TestCase{
			input:    "  @bot    r=popuko  ",
			expected: []string{"popuko"},
		},

		TestCase{
			input:    "@bot r=popuko-a",
			expected: []string{"popuko-a"},
		},
		TestCase{
			input:    "  @bot    r=popuko-a ",
			expected: []string{"popuko-a"},
		},

		TestCase{
			input:    "@bot r=popuko,pipimi",
			expected: []string{"popuko", "pipimi"},
		},
		TestCase{
			input:    "  @bot r=popuko,pipimi   ",
			expected: []string{"popuko", "pipimi"},
		},
		TestCase{
			input:    "  @bot r=popuko,  pipimi   ",
			expected: []string{"popuko", "pipimi"},
		},
		TestCase{
			input:    "  @bot r=popuko ,  pipimi   ",
			expected: []string{"popuko", "pipimi"},
		},
		TestCase{
			input:    "  @bot r= popuko ,  pipimi   ",
			expected: []string{"popuko", "pipimi"},
		},

		TestCase{
			input:    "@bot r=popuko-a,pipimi-b",
			expected: []string{"popuko-a", "pipimi-b"},
		},
		TestCase{
			input:    "  @bot r=popuko-a,pipimi-b   ",
			expected: []string{"popuko-a", "pipimi-b"},
		},
		TestCase{
			input:    "  @bot r=popuko-a,   pipimi-b   ",
			expected: []string{"popuko-a", "pipimi-b"},
		},
		TestCase{
			input:    "  @bot r=popuko-a  ,   pipimi-b   ",
			expected: []string{"popuko-a", "pipimi-b"},
		},
		TestCase{
			input:    "  @bot r= popuko-a  ,   pipimi-b   ",
			expected: []string{"popuko-a", "pipimi-b"},
		},
	}
	for _, testcase := range list {
		input := testcase.input

		ok, cmd := ParseCommand(input)
		if !ok {
			t.Errorf("input: `%v` should be ok", input)
			continue
		}

		v, ok := cmd.(*AcceptChangeByOthersCommand)
		if !ok {
			t.Errorf("input: `%v` should be AcceptChangeByOthersCommand", input)
			continue
		}

		if len(v.Reviewer) != len(testcase.expected) {
			t.Errorf("input: `%v` should be the expected length (`%v`) but the acutual length is `%v`", input, len(testcase.expected), len(v.Reviewer))
			continue
		}

		for i, actual := range v.Reviewer {
			expected := testcase.expected[i]
			if actual != expected {
				t.Errorf("input: `%v` should be the expected (`%v`) but `%v`", input, expected, actual)
				continue
			}
		}
	}
}

func TestParseCommandValidCaseForAssignReviewerCommand(t *testing.T) {
	type TestCase struct {
		input    string
		expected []string
	}

	list := []TestCase{
		TestCase{
			input:    "r? @reviewer",
			expected: []string{"reviewer"},
		},
		TestCase{
			input:    "r? @reviewer-a",
			expected: []string{"reviewer-a"},
		},
		TestCase{
			input:    "  r? @reviewer  ",
			expected: []string{"reviewer"},
		},
		TestCase{
			input:    "   r? @reviewer-a   ",
			expected: []string{"reviewer-a"},
		},

		TestCase{
			input:    "@reviewer r?",
			expected: []string{"reviewer"},
		},
		TestCase{
			input:    "@reviewer-a r?",
			expected: []string{"reviewer-a"},
		},
		TestCase{
			input:    "   @reviewer  r? ",
			expected: []string{"reviewer"},
		},
		TestCase{
			input:    "    @reviewer-a   r?",
			expected: []string{"reviewer-a"},
		},
	}
	for _, testcase := range list {
		input := testcase.input

		ok, cmd := ParseCommand(input)
		if !ok {
			t.Errorf("input: `%v` should be ok", input)
			continue
		}

		v, ok := cmd.(*AssignReviewerCommand)
		if !ok {
			t.Errorf("input: `%v` should be AssignReviewerCommand", input)
			continue
		}

		expected := testcase.expected[0]
		if actual := v.Reviewer; actual != expected {
			t.Errorf("input: `%v` should be the expected (`%v`) but `%v`", input, expected, actual)
			continue
		}
	}
}

func TestParseCommandValidCaseForCancelApprovedByReviewerCommand(t *testing.T) {
	type TestCase struct {
		input           string
		expectedBotName string
	}

	list := []TestCase{
		TestCase{
			input:           "@bot r-",
			expectedBotName: "bot",
		},
		TestCase{
			input:           "@bot-bot r-",
			expectedBotName: "bot-bot",
		},

		TestCase{
			input:           "    @bot r-",
			expectedBotName: "bot",
		},

		TestCase{
			input:           "@bot        r-",
			expectedBotName: "bot",
		},

		TestCase{
			input: `@bot        r-



	`,
			expectedBotName: "bot",
		},
	}
	for _, testcase := range list {
		input := testcase.input

		ok, cmd := ParseCommand(input)
		if !ok {
			t.Errorf("input: `%v` should be ok", input)
			continue
		}

		v, ok := cmd.(*CancelApprovedByReviewerCommand)
		if !ok {
			t.Errorf("input: `%v` should be CancelApprovedByReviewerCommand", input)
			continue
		}

		expected := testcase.expectedBotName
		if actual := v.BotName(); actual != expected {
			t.Errorf("input: `%v` should be the expected bot (`%v`) name but `%v`", input, expected, actual)
			continue
		}
	}
}

func TestParseCommandInvalidCase(t *testing.T) {
	input := []string{
		"Hello, I'm john.",
		"",
		"bot r+",
		"@bot",

		// r+
		"@bot r +",
		"@bot r r+",
		"@bot r+ r",
		" @ bot r+",
		" @ bot r +",
		`
    @bot r+`,
		`@bot
    r+`,

		// r-
		"@bot r -",
		"@bot r r-",
		"@bot r- r",
		" @ bot r-",
		" @ bot r -",
		`
    @bot r-`,
		`@bot
    r-`,

		// r=reviewer
		"@bot r =a",
		"@bot r = a",
		"@bot r r=a",
		"@bot r=a r",
		" @ bot r=a",
		" @ bot r = a",
		" @ bot r =a",
		`
    @bot r=a`,
		`@bot
    r=a`,

		// @reviewer r?
		"@bot r r?",
		"@bot r? r",
		"@bot r? @bot2",
		"@bot r ?",
		" @ bot r?",
		" @ bot r ? ",
		`
    @bot r?`,
		`@bot
    r?`,

		// r? @reviewer
		"r? r @bot",
		"r? @bot r",
		"r? @bot r @bot2",
		"r ? @bot",
		" r? @ bot",
		" r ? @ bot ",
		`
    r? @bot`,
		`r?
    @bot`,
	}
	for _, item := range input {
		if ok, _ := ParseCommand(item); ok {
			t.Errorf("%v should not be ok", item)
		}
	}
}
