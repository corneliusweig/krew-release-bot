package main

import "testing"

func TestHandler(t *testing.T) {
	testcases := []struct {
		name     string
		request  string
		response string
	}{
		{},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}
