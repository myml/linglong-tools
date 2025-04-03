package cmd

import "testing"

func Test_preSignDiectory(t *testing.T) {
	err := preSignDiectory("")
	if err != nil {
		t.Error(err)
	}
}
