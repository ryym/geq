package tests

import "testing"

type testCase struct {
	name string
	run  func() error
}

func runTestCases(t *testing.T, cases []testCase) {
	for i, c := range cases {
		err := c.run()
		if err != nil {
			t.Logf("failed: case[%d] %s", i, c.name)
			t.Error(err)
		}
	}
}
