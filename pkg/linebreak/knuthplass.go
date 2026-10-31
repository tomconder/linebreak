package linebreak

// KnuthPlass is a temporary alias for MinRaggedness. The true Knuth-Plass
// algorithm replaces it.
func KnuthPlass(text string, width int) []string {
	return MinRaggedness(text, width)
}
