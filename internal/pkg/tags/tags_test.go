package tags

import "testing"

func TestMatch(t *testing.T) {
	testCases := []struct {
		name              string
		tag               string
		tagPatternInclude []string
		tagPatternExclude []string
		expect            bool
		expectErr         error
	}{
		{
			"simple test, single include, no patterns, match",
			"tag1",
			[]string{"tag1"},
			[]string{},
			true,
			nil,
		},
		{
			"simple test, multiple include, no patterns, match",
			"tag4",
			[]string{"tag1", "tag2", "tag3", "tag4"},
			[]string{},
			true,
			nil,
		},
		{
			"matching include & exlude",
			"tag4",
			[]string{"tag1", "tag2", "tag3", "tag4"},
			[]string{"tag2", "tag4"},
			false,
			nil,
		},
		{
			"matching include pattern, no excludes",
			"tag4",
			[]string{"tag*"},
			[]string{},
			true,
			nil,
		},
		{
			"matching include pattern, match exclude pattern",
			"tag4",
			[]string{"tag*"},
			[]string{"*ag*"},
			false,
			nil,
		},
		{
			"matching include pattern, not match exclude pattern",
			"tag4",
			[]string{"tag1", "tag2", "tag3", "tag*"},
			[]string{"tag1", "tag2"},
			true,
			nil,
		},
		{
			"matching include pattern, match exclude pattern",
			"tag4",
			[]string{"tag1", "tag2", "tag3", "tag*"},
			[]string{"tag1", "*"},
			false,
			nil,
		},
		{
			"match >= in include pattern",
			"v0.4",
			[]string{">=v0.1.2"},
			[]string{},
			true,
			nil,
		},
		{
			"not match >= in include pattern",
			"v0.1",
			[]string{">=v0.1.2"},
			[]string{},
			false,
			nil,
		},
		{
			"exact match >= in include pattern",
			"v0.1",
			[]string{">=v0.1"},
			[]string{},
			true,
			nil,
		},
		{
			"exact match >= in include pattern",
			"v0.1.5",
			[]string{">=v0.1.1"},
			[]string{"v0.1.3"},
			true,
			nil,
		},
		{
			"exact match >= in include pattern",
			"v0.2.5",
			[]string{">=v0.1.1"},
			[]string{"v0.2.*"},
			false,
			nil,
		},
		{
			"special regexp characters",
			"v111",
			[]string{"v1.1"},
			[]string{},
			false,
			nil,
		},
	}

	for _, testCase := range testCases {
		match, err := Match(testCase.tag, testCase.tagPatternInclude, testCase.tagPatternExclude)
		if err != testCase.expectErr {
			t.Errorf("test %s failed, expected err to be: %s", testCase.name, testCase.expectErr)
		}
		if match != testCase.expect {
			t.Errorf("test %s failed, expected %t match", testCase.name, testCase.expect)
		}
	}
}
