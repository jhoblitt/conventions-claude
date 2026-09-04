package audit

// LooksLikeLdflagsX exposes the -X search to the spec, which drives it with the
// command lines it must and must not read as a version stamp.
func LooksLikeLdflagsX(line string) bool { return ldflagsX.MatchString(line) }
