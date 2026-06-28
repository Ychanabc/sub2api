package service

func derefStringPtr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func derefIntPtr(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
