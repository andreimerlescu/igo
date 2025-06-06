package internal

type ErrOSNotSupported struct {
	OS string
}

func (e ErrOSNotSupported) Error() string {
	return "igo does not support " + e.OS + " yet"
}

type ErrBitNotSet struct {
	Err error
}

func (e ErrBitNotSet) Error() string {
	return "failed to set setuid/setgid bits: " + e.Err.Error()
}

type ErrPathFailed struct {
	Path string
	Err  error
	Neg  string
}

func (e ErrPathFailed) Error() string {
	return "error doing " + e.Neg + " symlink " + e.Path + ": " + e.Err.Error()
}

type ErrChmodFailed struct {
	Path string
	Err  error
}

func (e ErrChmodFailed) Error() string {
	return "failed to chmod u+w on " + e.Path + ": " + e.Err.Error()
}

type ErrStickyBitsOnFile struct {
	Path string
}

func (e ErrStickyBitsOnFile) Error() string {
	return "setting sticky bits from files do nothing: " + e.Path
}

type ErrStickyBitFailed struct {
	Path string
	Err  error
	How  string
}

func (e ErrStickyBitFailed) Error() string {
	return "failed to " + e.How + " sticky bit: " + e.Err.Error()
}

type ErrSetUIDGIDBit struct {
	Path string
	Err  error
	How  string
}

func (e ErrSetUIDGIDBit) Error() string {
	return "tried to " + e.How + " setuid/setgid bits: " + e.Path
}

type ErrFile struct {
	Path string
	Err  error
	How  string
}

func (e ErrFile) Error() string {
	return e.How + " " + e.Path + " yields: " + e.Err.Error()
}

type ErrDirEntries struct {
	Path string
	Err  error
}

func (e ErrDirEntries) Error() string {
	return "failed to read directory entries: " + e.Err.Error()
}
