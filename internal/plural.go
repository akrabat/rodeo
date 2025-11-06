package internal

func PluralS(s interface{}) string {
	length := 1
	switch v := s.(type) {
	case int:
		length = v
	case []string:
		length = len(v)
	case []Album:
		length = len(v)
	}
	if length == 1 {
		return ""
	}
	return "s"
}
