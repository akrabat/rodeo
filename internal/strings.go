package internal

import "strings"

// from https://stackoverflow.com/a/45024302/23060
func filter(src []string) (res []string) {
	for _, s := range src {
		newStr := strings.Join(res, " ")
		if !strings.Contains(newStr, s) {
			res = append(res, s)
		}
	}
	return
}

// from https://stackoverflow.com/a/45024302/23060
func Intersection(section1, section2 []string) (intersection []string) {
	str1 := strings.Join(filter(section1), " ")
	for _, s := range filter(section2) {
		if strings.Contains(str1, s) {
			intersection = append(intersection, s)
		}
	}
	return
}


// From: https://stackoverflow.com/a/56377243/23060
func Difference(a, b []string) []string {
	temp := map[string]int{}
	for _, s := range a {
		temp[s]++
	}
	for _, s := range b {
		temp[s]--
	}

	var result []string
	for s, v := range temp {
		if v != 0 {
			result = append(result, s)
		}
	}
	return result
}
