package keys

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// maxFileSize bounds what is read from one file or stream (64 MiB).
const maxFileSize = 64 << 20

// StdinName is the path argument that selects standard input.
const StdinName = "-"

var keyExtensions = map[string]bool{
	".pem": true, ".crt": true, ".cer": true, ".der": true,
	".key": true, ".pub": true, ".txt": true,
}

// Load reads every path: a file is parsed whatever its extension, a directory
// is walked recursively and only files with a key/certificate extension are
// read. The path "-" reads stdin. Unreadable inputs end up in Result.Errors so
// one bad file does not stop an audit.
func Load(paths []string, stdin io.Reader) Result {
	var res Result
	for _, p := range paths {
		if p == StdinName {
			res.merge(loadStream(stdin, "stdin"))
			continue
		}
		info, err := os.Stat(p)
		if err != nil {
			res.Errors = append(res.Errors, err)
			continue
		}
		if info.IsDir() {
			res.merge(loadDir(p))
			continue
		}
		res.merge(loadFile(p))
	}
	return res
}

func loadDir(root string) Result {
	var res Result
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			res.Errors = append(res.Errors, err)
			return nil
		}
		if d.IsDir() || !keyExtensions[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		res.merge(loadFile(path))
		return nil
	})
	if err != nil {
		res.Errors = append(res.Errors, err)
	}
	return res
}

func loadFile(path string) Result {
	f, err := os.Open(path) //nolint:gosec // the path is the user's explicit input
	if err != nil {
		return Result{Errors: []error{err}}
	}
	defer func() { _ = f.Close() }()
	return loadStream(f, filepath.ToSlash(path))
}

func loadStream(r io.Reader, source string) Result {
	data, err := io.ReadAll(io.LimitReader(r, maxFileSize+1))
	if err != nil {
		return Result{Errors: []error{fmt.Errorf("%s: %w", source, err)}}
	}
	if len(data) > maxFileSize {
		return Result{Errors: []error{fmt.Errorf("%s: larger than %d MiB", source, maxFileSize>>20)}}
	}
	return ParseBytes(data, source)
}
