package tests

import "testing"

type testCase struct {
	name string
	run  func() bool
}

func runTestCases(t *testing.T, cases []testCase) {
	for i, c := range cases {
		ok := c.run()
		if !ok {
			t.Logf("failed: case[%d] %s", i, c.name)
		}
	}
}
