// Package finder allows for retrieving a list of Maildirs beneath a
// given prefix, and the messages contained within them.
//
// It is a very loose wrapper around `path/filepath` with no particularly
// special logic.
package finder

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Finder holds our state.
type Finder struct {

	// Prefix is the root directory of the users' maildir hierarchy.
	Prefix string
}

// New creates a new object which can be used to find Maildir folders,
// and the messages contained within them.
func New(prefix string) *Finder {
	return &Finder{Prefix: prefix}
}

// Messages returns all message-files beneath the given maildir folder.
//
// This means we walk the filesystem returning the list of filenames present
// beneath $path/new/ and $path/cur
//
// We exclude non-files, and ignore $path/tmp/.
func (f *Finder) Messages(path string) []string {

	// TODO use a struct
	//    filename string
	//    info os.FileInfo
	//
	//Then we can sort them like so:
	//
	// sort.Slice(files, func(i,j int) bool{
	//	return files[i].ModTime().Unix() < files[j].ModTime().Unix()
	// })
	//
	// Doing this is better than getting all the info a second
	// time, once we've found the filenames
	//

	// Build the list of message filenames here.
	var files []string

	// Directories we examine beneath the maildir
	dirs := []string{"cur", "new"}

	// For each subdirectory
	for _, dir := range dirs {

		// Build up the complete-path
		prefix := filepath.Join(path, dir)

		// Now record all files beneath that directory
		_ = filepath.Walk(prefix, func(path string, f os.FileInfo, err error) error {

			// We only care about files
			mode := f.Mode()
			if mode.IsRegular() {
				files = append(files, path)
			}

			return nil
		})
	}

	return files
}

// Maildirs returns the list of Maildir folders beneath our prefix.
//
// This function handles recursive/nested maildir folders.
func (f *Finder) Maildirs() []string {

	maildirs := []string{}

	// Subdirectories we care about
	dirs := []string{"cur", "new", "tmp"}

	//
	// Find maildirs
	//
	_ = filepath.Walk(f.Prefix, func(path string, f os.FileInfo, err error) error {

		// Ignore the new/cur/tmp directories we'll terminate with
		for _, dir := range dirs {
			if strings.HasSuffix(path, dir) {
				return nil
			}
		}

		//
		// Ignore non-directories. (We might might find Dovecot
		// index-files, etc.)
		//
		mode := f.Mode()
		if !mode.IsDir() {
			return nil
		}

		//
		// Look for `new/`, `cur/`, and `tmp/` subdirectoires
		// beneath the given directory.
		//
		// If any are missing then this is not a maildir.
		//
		for _, dir := range dirs {
			path := filepath.Join(path, dir)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return nil
			}
		}

		//
		// Found a positive result
		//
		maildirs = append(maildirs, path)
		return nil
	})

	//
	// Sort
	//
	sort.Slice(maildirs, func(i, j int) bool { return strings.ToLower(maildirs[i]) < strings.ToLower(maildirs[j]) })

	return maildirs
}
