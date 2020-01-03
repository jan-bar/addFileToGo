package LzmaSpec

// from 7zTypes.h
const (
	SzOk               = 0
	SzErrorData        = 1
	SzErrorMem         = 2
	SzErrorCrc         = 3
	SzErrorUnsupported = 4
	SzErrorParam       = 5
	SzErrorInputEOF    = 6
	SzErrorOutputEOF   = 7
	SzErrorRead        = 8
	SzErrorWrite       = 9
	SzErrorProgress    = 10
	SzErrorFail        = 11
	SzErrorThread      = 12
	SzErrorArchive     = 16
	SzErrorNoArchive   = 17
)

const (
	LzmaPropsSize = uint64(5)

	sizeHeader = 8
)
