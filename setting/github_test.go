package setting

import "testing"

const expectedK = "bar"
const expectedV = "foo"

func TestGithubSettingAccept1(t *testing.T) {
	s := GithubSetting{
		Repositories: make([]string, 0),
	}
	initGithubSetting(&s)

	if actual := s.accept(expectedK, expectedV); !actual {
		t.Fatalf("expected: %v\n", actual)
	}
}

func TestGithubSettingAccept2(t *testing.T) {
	s := GithubSetting{
		Repositories: nil,
	}
	initGithubSetting(&s)

	if actual := s.accept(expectedK, expectedV); !actual {
		t.Fatalf("expected: %v\n", actual)
	}
}

func TestGithubSettingAccept3(t *testing.T) {
	s := GithubSetting{
		Repositories: []string{expectedK + "/" + expectedV},
	}
	initGithubSetting(&s)

	if actual := s.accept(expectedK, expectedV); !actual {
		t.Fatalf("expected: %v\n", actual)
	}
}

func TestGithubSettingAccept4(t *testing.T) {
	s := GithubSetting{
		Repositories: []string{"notbar/notfoo"},
	}
	initGithubSetting(&s)

	if actual := s.accept(expectedK, expectedV); actual {
		t.Fatalf("expected: %v\n", actual)
	}
}
