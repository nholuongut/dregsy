package sync

import "testing"

func TestIsValidTag(t *testing.T) {
	for _, testCase := range []struct {
		tag   string
		valid bool
	}{
		{
			tag:   "tag1",
			valid: true,
		},
		{
			tag:   ">=ẗag1*",
			valid: false,
		},
		{
			tag:   "tag#$%",
			valid: false,
		},
		{
			tag:   "v1.2.*-1",
			valid: true,
		},
		{
			tag:   ">=v1.2.1",
			valid: true,
		},
		{
			tag:   ">= v1.2.1", // no space allowed
			valid: false,
		},
		{
			tag:   ">=v1.2.*", // no wildcard when using comparison operators
			valid: false,
		},
	} {
		err := isValidTag(testCase.tag)
		if testCase.valid && err != nil {
			t.Errorf("failed validation of tag %s - should be valid, got %s", testCase.tag, err)
		}
		if !testCase.valid && err == nil {
			t.Errorf("failed validation of tag %s - should be invalid", testCase.tag)
		}
	}
}
