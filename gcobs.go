package gcobs

import (
	"errors"
)

var (
	ErrorOverflow                 = errors.New("overflow")
	ErrorEmptySrc                 = errors.New("empty src")
	ErrorInvalidCobsFrameWithZero = errors.New("invalid cobs frame with zero byte")
	ErrorInvalidCobsFrameTooShort = errors.New("invalid cobs frame ,too short")
	ErrorInvalidCobsFrameTooLong  = errors.New("invalid cobs frame ,too long")
)

// encodeDstBufLenMax calculates the maximum length of the destination buffer
func encodeDstBufLenMax(srcLen int) int {
	if srcLen == 0 {
		return 1
	}
	return srcLen + (srcLen+253)/254
}

// decodeDstBufLenMax calculates the maximum length of the destination buffer
func decodeDstBufLenMax(srcLen int) int {
	if srcLen == 0 {
		return 1
	}
	return srcLen - 1
}

// Encode encodes the input byte slice using COBS (Consistent Overhead Byte Stuffing) encoding.
func Encode(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, ErrorEmptySrc
	}
	dstBuffLen := encodeDstBufLenMax(len(src))
	dstBuff := make([]byte, dstBuffLen)

	idxSrcRead := 0
	idxSrcEnd := len(src)

	idxDstCodeWrite := 0
	idxDstDataWrite := 1
	idxDstEnd := dstBuffLen

	codeSearchLen := uint8(1)
	for {
		// Check for running out of output buffer space
		if idxDstDataWrite >= idxDstEnd {
			return nil, ErrorOverflow
		}
		srcByte := src[idxSrcRead]
		idxSrcRead++

		if srcByte == 0 {
			//  We found a zero byte
			dstBuff[idxDstCodeWrite] = codeSearchLen
			idxDstCodeWrite = idxDstDataWrite
			idxDstDataWrite++

			codeSearchLen = 1
			if idxSrcRead >= idxSrcEnd {
				break
			}
		} else {
			dstBuff[idxDstDataWrite] = srcByte
			idxDstDataWrite++
			codeSearchLen++
			if idxSrcRead >= idxSrcEnd {
				break
			}
			if codeSearchLen == 0xFF {
				/* We have a long string of non-zero bytes, so we need
				 * to write out a length code of 0xFF. */
				dstBuff[idxDstCodeWrite] = codeSearchLen
				idxDstCodeWrite = idxDstDataWrite
				idxDstDataWrite++
				codeSearchLen = 1
			}
		}
	}
	/* We've reached the end of the source data (or possibly run out of output buffer)
	 * Finalise the remaining output. In particular, write the code (length) byte.
	 * Update the pointer to calculate the final output length.
	 */
	if idxDstCodeWrite >= idxDstEnd {
		/* We've run out of output buffer to write the code byte. */
		return nil, ErrorOverflow
	} else {
		/* Write the last code (length) byte. */
		dstBuff[idxDstCodeWrite] = codeSearchLen
	}
	return dstBuff[:idxDstDataWrite], nil
}

// EncodeWitZeroEnd encodes the input byte slice using COBS  encoding and appends a zero byte at the end.
func EncodeWitZeroEnd(src []byte) ([]byte, error) {
	if encode, err := Encode(src); err != nil {
		return nil, err
	} else {
		return append(encode, 0x00), nil
	}
}

// Decode decodes the input byte slice using COBS (Consistent Overhead Byte Stuffing) decoding.
func Decode(src []byte) ([]byte, error) {
	dstBuffLen := decodeDstBufLenMax(len(src))
	dstBuff := make([]byte, dstBuffLen)

	idxDstDataWrite := 0
	idxDstEnd := dstBuffLen

	idxSrcRead := 0
	idxSrcEnd := len(src)

	for {
		lenCode := src[idxSrcRead]
		idxSrcRead++

		if lenCode == 0 {
			return nil, ErrorInvalidCobsFrameWithZero
		}

		lenCode--
		//  Check length code against remaining input bytes
		remainBytes := idxSrcEnd - idxSrcRead
		if int(lenCode) > remainBytes {
			return nil, ErrorInvalidCobsFrameTooShort
		}

		remainBytes = idxDstEnd - idxDstDataWrite
		// Check length code against remaining output buffer space
		if int(lenCode) > remainBytes {
			return nil, ErrorInvalidCobsFrameTooLong
		}

		for i := lenCode; i != 0; i-- {
			srcByte := src[idxSrcRead]
			idxSrcRead++
			if srcByte == 0 {
				return nil, ErrorInvalidCobsFrameWithZero
			}
			dstBuff[idxDstDataWrite] = srcByte
			idxDstDataWrite++
		}
		if idxSrcRead >= idxSrcEnd {
			break
		}
		// Add a zero to the end
		if lenCode != 0xFE {
			if idxDstDataWrite >= idxDstEnd {
				return nil, ErrorOverflow
			}
			dstBuff[idxDstDataWrite] = 0
			idxDstDataWrite++
		}
	}
	return dstBuff[:idxDstDataWrite], nil
}

// DecodeWithZeroEndIf decodes the input byte slice using COBS decoding and removes the trailing zero byte if present.
func DecodeWithZeroEndIf(src []byte) ([]byte, error) {
	if src[len(src)-1] == 0x00 {
		src = src[:len(src)-1]
	}
	return Decode(src)
}
