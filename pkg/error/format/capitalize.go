package format

// CapitalizeError formats error messages for JSON responses.
func CapitalizeError(err error) string {
	msg := err.Error()
	if len(msg) == 0 {
		return msg
	}
	return string(msg[0]-32) + msg[1:] + "."
}
