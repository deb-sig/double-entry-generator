package util

import "strings"

func SplitFindContains(str, target, sep string, match bool) bool {
	ss := strings.Split(str, sep)
	isContain := false
	for _, s := range ss {
		if strings.Contains(target, s) {
			isContain = true
			break
		}
	}
	if !isContain {
		return false
	}
	return match
}
