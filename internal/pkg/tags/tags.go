package tags

import (
	"fmt"
	"regexp"
	"strings"
)

func GetComparisonOperator(tag string) string {
	if strings.HasPrefix(tag, "<=") {
		return "<="
	}
	if strings.HasPrefix(tag, "<") {
		return "<"
	}
	if strings.HasPrefix(tag, ">=") {
		return ">="
	}
	if strings.HasPrefix(tag, ">") {
		return ">"
	}

	return ""
}

func patternToRegexp(pattern string) (*regexp.Regexp, error) {
	pattern = regexp.QuoteMeta(pattern)
	return regexp.Compile(strings.ReplaceAll(pattern, "\\*", "[0-9a-zA-Z-_]*"))
}

func matchPattern(tag, pattern string) (bool, error) {
	switch op := GetComparisonOperator(pattern); op {
	case "<=":
		return strings.Compare(tag, pattern[2:]) <= 0, nil

	case "<":
		return strings.Compare(tag, pattern[1:]) < 0, nil

	case ">=":
		return strings.Compare(tag, pattern[2:]) >= 0, nil

	case ">":
		return strings.Compare(tag, pattern[1:]) > 0, nil

	case "":

	default:
		return false, fmt.Errorf("unexpected comparison operator found in tag %s", tag)

	}

	patternReg, err := patternToRegexp(pattern)
	if err != nil {
		return false, err
	}

	if patternReg.Match([]byte(tag)) {
		return true, nil
	}

	return false, nil
}

func Match(tag string, tagPatternInclude []string, tagPatternExclude []string) (bool, error) {
	for _, exclude := range tagPatternExclude {
		match, err := matchPattern(tag, exclude)
		if err != nil || match == true {
			return false, err
		}
	}

	for _, include := range tagPatternInclude {
		match, err := matchPattern(tag, include)
		if err != nil || match == true {
			return true, err
		}
	}

	return false, nil
}
