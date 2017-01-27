package queue

import "testing"

func Test_createAbs_ValidCase(t *testing.T) {
	type Testcase struct {
		root     string
		path     string
		expected string
	}
	list := []Testcase{
		Testcase{
			root:     "/a",
			path:     "b",
			expected: "/a/b",
		},
		Testcase{
			root:     "/a",
			path:     "/b",
			expected: "/a/b",
		},
		Testcase{
			root:     "/a",
			path:     ".b",
			expected: "/a/.b",
		},
		Testcase{
			root:     "/a",
			path:     "../~/b",
			expected: "/a/~/b",
		},
		Testcase{
			root:     "/a",
			path:     "./b",
			expected: "/a/b",
		},
		Testcase{
			root:     "/a",
			path:     "../../b",
			expected: "/a/b",
		},
		Testcase{
			root:     "/a",
			path:     "..",
			expected: "/a",
		},
		Testcase{
			root:     "..",
			path:     "/a",
			expected: "/a",
		},
		Testcase{
			root:     "a",
			path:     "/b",
			expected: "/a/b",
		},
		Testcase{
			root:     "a",
			path:     "../b",
			expected: "/a/b",
		},
		Testcase{
			root:     "a",
			path:     "../~/b",
			expected: "/a/~/b",
		},
	}

	for _, test := range list {
		abs, err := createAbs(test.root, test.path)
		if abs != test.expected {
			t.Errorf("%+v should be `%v`, but the acutual is `%v`", test, test.expected, abs)
		}
		if err != nil {
			t.Errorf("%+v should not return `err` but %v", test, err)
		}
	}
}

func Test_createAbs_InvalidCase(t *testing.T) {
	type Testcase struct {
		root string
		path string
	}
	list := []Testcase{
		Testcase{
			root: "",
			path: "",
		},
		Testcase{
			root: "/a",
			path: "",
		},
		Testcase{
			root: "/",
			path: "",
		},
		Testcase{
			root: "",
			path: "/",
		},
		Testcase{
			root: "/",
			path: "/",
		},
		Testcase{
			root: ".",
			path: ".",
		},
		Testcase{
			root: ".",
			path: "..",
		},
		Testcase{
			root: "..",
			path: ".",
		},
		Testcase{
			root: "..",
			path: "..",
		},
	}

	for _, test := range list {
		abs, err := createAbs(test.root, test.path)
		if abs != "" {
			t.Errorf("%+v should be empty string, but the acutual is `%v`", test, abs)
		}

		if err == nil {
			t.Errorf("%+v should not be nil", test)
		}
	}
}
