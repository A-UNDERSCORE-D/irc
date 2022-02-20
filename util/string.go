package util

func min(a, b int) int {
	if a > b {
		return b
	}

	return a
}

// ChunkMessage returns the given message in chunks with a max size of chunkLength
func ChunkMessage(message string, chunkLength int) []string {
	if chunkLength <= 0 {
		panic("ChunkMessage: chunk size <= 0")
	}
	if len(message) <= chunkLength {
		return []string{message}
	}

	out := []string{}

	for len(message) > chunkLength {
		out = append(out, message[:min(chunkLength, len(message))])
		message = message[min(chunkLength, len(message)):]
	}

	out = append(out, message)

	return out
}
