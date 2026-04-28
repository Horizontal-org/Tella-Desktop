package genericutil

import (
	"os"
)

const USER_ONLY_FILE_PERMS = 0600
const USER_ONLY_DIR_PERMS = 0700

// NarrowCreate creates a file with narrow access permissions (limited to the current user).
// NOTE: Does not close file with defer! Caller should call as:
//
// f, err := NarrowCreate(filepath)
// if err != nil { /* error handling / escape */ }
// defer f.Close()

func NarrowCreate(fpath string) (*os.File, error) {
	file, err := os.Create(fpath)
	if err != nil {
		return nil, err
	}
	err = file.Chmod(USER_ONLY_FILE_PERMS)
	if err != nil {
		closeErr := file.Close()
		if closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}
	return file, nil
}

// Routine has been lifted from matthewhartstonge/argon2 so as to not include all of argon2 in packages where that is
// not needed. License: https://github.com/matthewhartstonge/argon2?tab=Apache-2.0-1-ov-file#readme
//
// SecureZeroMemory is a helper method which sets all bytes in `b`
// (up to its capacity) to `0x00`, erasing its contents.
func SecureZeroMemory(b []byte) {
	b = b[:cap(b):cap(b)]
	for i := range b {
		b[i] = 0
	}
}
