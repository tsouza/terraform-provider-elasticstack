package provider

func expandStringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, v.(string))
		}
	}
	return vs
}

func collapseStringList(list []string) []interface{} {
	var collapsed []interface{}
	for _, i := range list {
		collapsed = append(collapsed, i)
	}
	return collapsed
}
