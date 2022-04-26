package gmo

//converts map[string]string to query string format
// ?key=val&keyN=valN

func Querystr(param map[string]string) string {
	if param == nil {
		return ""
	}
	query := "?" + querystr(param)
	return query
}

func querystr(param map[string]string) string {
	query := ""
	for k, v := range param {
		query += k + "=" + v + "&"
	}
	return query[:len(query)-1]
}
