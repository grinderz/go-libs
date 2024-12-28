package libbytes

func JoinWithAlloc(size int, s ...[]byte) []byte {
	buf, index := make([]byte, size), 0
	for _, v := range s {
		index += copy(buf[index:], v)
	}

	return buf
}
