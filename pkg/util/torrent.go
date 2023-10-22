package util

// GetLengthForIdx calculates the length of a piece depending on its index
func GetLengthForIdx(tLen, pieceLen, idx int) int {

	noOfPieces := int((tLen + pieceLen - 1) / pieceLen)

	// if last
	if idx == noOfPieces-1 {
		rem := tLen % pieceLen
		// but even division
		if rem == 0 {
			return pieceLen
		}
		return rem
	}

	return pieceLen
}
