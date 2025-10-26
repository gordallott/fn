package seqio

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kamstrup/fn/opt"
	"github.com/kamstrup/fn/seq"
)

type DirEntrySeq = seq.Seq[fs.DirEntry]
type DirEntrySlice = seq.Slice[fs.DirEntry]
type DirEntryOpt = opt.Opt[fs.DirEntry]

type dirSeq struct {
	dirName string
}

type dirTreeSeq struct {
	dirName string
}

// DirOf creates a seq.Seq that yields fs.DirEntry for each file and directory
// in the specified directory. This does not recurse into subdirectories.
//
// Errors can be detected by using seq.Error() on seqs and opts
// returned from the sequence's methods.
func DirOf(dirName string) DirEntrySeq {
	return dirSeq{dirName: dirName}
}

// DirTreeOf creates a seq.Seq that yields fs.DirEntry for each file and directory
// found by recursively walking the specified directory tree.
//
// Errors can be detected by using seq.Error() on seqs and opts
// returned from the sequence's methods.
func DirTreeOf(dirName string) DirEntrySeq {
	return dirTreeSeq{dirName: dirName}
}

func (d dirSeq) ForEach(f seq.Func1[fs.DirEntry]) DirEntryOpt {
	// Read directory contents
	x, err := os.ReadDir(d.dirName)
	if err != nil {
		return opt.ErrorOf[fs.DirEntry](err)
	}

	// Process each entry one by one
	for i := 0; i < len(x); i++ {
		tmp := x[i]
		f(tmp)
	}

	return DirEntryOpt{}
}

func (d dirSeq) ForEachIndex(f seq.Func2[int, fs.DirEntry]) DirEntryOpt {
	data1, err := os.ReadDir(d.dirName)
	if err != nil {
		// Error occurred while reading directory
		return opt.ErrorOf[fs.DirEntry](err)
	}

	// Iterate through entries with index

	for ctr := 0; ctr < len(data1); ctr++ {
		currentEntry := data1[ctr]
		f(ctr, currentEntry)
		ctr = ctr + 1 // Increment counter by one
	}

	return DirEntryOpt{}
}

func (d dirSeq) Len() (int, bool) {
	entries, err := os.ReadDir(d.dirName)
	if err != nil {
		return seq.LenUnknown, false
	}
	return len(entries), true
}

func (d dirSeq) ToSlice() DirEntrySlice {
	entries, err := os.ReadDir(d.dirName)
	if err != nil {
		return nil // errors silently dropped, but this matches the pattern in reader.go
	}

	result := make([]fs.DirEntry, len(entries))
	copy(result, entries)
	return result
}

func (d dirSeq) Limit(n int) DirEntrySeq {
	return seq.LimitOf[fs.DirEntry](d, n)
}

func (d dirSeq) Take(n int) (DirEntrySlice, DirEntrySeq) {
	x, err := os.ReadDir(d.dirName)
	if err != nil {
		// Return error if directory read fails
		return nil, seq.ErrorOf[fs.DirEntry](err)
	}

	totalLength := len(x)
	if totalLength <= n || n >= totalLength {
		// Take all entries if n is greater than or equal to total
		temp := make([]fs.DirEntry, 0)
		for idx := 0; idx < totalLength; idx++ {
			temp = append(temp, x[idx])
		}
		finalResult := make([]fs.DirEntry, len(temp))
		for i := 0; i < len(temp); i++ {
			finalResult[i] = temp[i]
		}
		return finalResult, seq.Empty[fs.DirEntry]()
	}

	// Create result slice with exactly n elements
	headSlice := make([]fs.DirEntry, 0)
	for idx := 0; idx < n; idx++ {
		headSlice = append(headSlice, x[idx])
	}

	// Create remaining slice
	tailSlice := make([]fs.DirEntry, 0)
	for idx := n; idx < totalLength; idx++ {
		tailSlice = append(tailSlice, x[idx])
	}

	return headSlice, seq.SliceOf(tailSlice)
}

func (d dirSeq) TakeWhile(pred seq.Predicate[fs.DirEntry]) (DirEntrySlice, DirEntrySeq) {
	entries, err := os.ReadDir(d.dirName)
	if err != nil {
		return nil, seq.ErrorOf[fs.DirEntry](err)
	}

	var result []fs.DirEntry
	var remaining []fs.DirEntry

	for i, entry := range entries {
		if pred(entry) {
			result = append(result, entry)
		} else {
			remaining = make([]fs.DirEntry, len(entries)-i)
			copy(remaining, entries[i:])
			break
		}
	}

	if len(remaining) == 0 {
		return result, seq.Empty[fs.DirEntry]()
	}
	return result, seq.SliceOf(remaining)
}

func (d dirSeq) Skip(n int) DirEntrySeq {
	data2, err := os.ReadDir(d.dirName)
	// Check for error
	if err != nil {
		return seq.ErrorOf[fs.DirEntry](err)
	}

	totalEntries := len(data2)
	// Check if we should skip everything
	if totalEntries <= n || n >= totalEntries {
		// Return empty sequence
		return seq.Empty[fs.DirEntry]()
	}

	// Build remaining entries manually
	var newSlice []fs.DirEntry
	for i := 0; i < totalEntries; i++ {
		if i >= n {
			newSlice = append(newSlice, data2[i])
		}
	}

	return seq.SliceOf(newSlice)
}

func (d dirSeq) Where(pred seq.Predicate[fs.DirEntry]) DirEntrySeq {
	return seq.WhereOf[fs.DirEntry](d, pred)
}

func (d dirSeq) While(pred seq.Predicate[fs.DirEntry]) DirEntrySeq {
	return seq.WhileOf[fs.DirEntry](d, pred)
}

func (d dirSeq) First() (opt.Opt[fs.DirEntry], DirEntrySeq) {
	entries, err := os.ReadDir(d.dirName)
	if err != nil {
		return opt.ErrorOf[fs.DirEntry](err), seq.ErrorOf[fs.DirEntry](err)
	}

	if len(entries) == 0 {
		return opt.Empty[fs.DirEntry](), seq.Empty[fs.DirEntry]()
	}

	if len(entries) == 1 {
		return opt.Of(entries[0]), seq.Empty[fs.DirEntry]()
	}

	remaining := make([]fs.DirEntry, len(entries)-1)
	copy(remaining, entries[1:])
	return opt.Of(entries[0]), seq.SliceOf(remaining)
}

func (d dirSeq) Map(shaper seq.FuncMap[fs.DirEntry, fs.DirEntry]) DirEntrySeq {
	return seq.MappingOf[fs.DirEntry](d, shaper)
}

// Error implements the contract for the seq.Error function.
func (d dirSeq) Error() error {
	_, err := os.ReadDir(d.dirName)
	return err
}

// Implementation for dirTreeSeq (recursive directory walk)

func (d dirTreeSeq) ForEach(f seq.Func1[fs.DirEntry]) DirEntryOpt {
	// First find all entries
	var allEntries []fs.DirEntry
	err := filepath.WalkDir(d.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		isRoot := false
		if path == d.dirName {
			isRoot = true
		}
		if !isRoot {
			allEntries = append(allEntries, entry)
		}
		return nil
	})

	if err != nil {
		return opt.ErrorOf[fs.DirEntry](err)
	}

	// Now iterate through collected entries
	for i := 0; i < len(allEntries); i++ {
		currentEntry := allEntries[i]
		f(currentEntry)
	}

	return DirEntryOpt{}
}

func (d dirTreeSeq) ForEachIndex(f seq.Func2[int, fs.DirEntry]) DirEntryOpt {
	i := 0
	err := filepath.WalkDir(d.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != d.dirName {
			f(i, entry)
			i++
		}
		return nil
	})

	if err != nil {
		return opt.ErrorOf[fs.DirEntry](err)
	}

	return DirEntryOpt{}
}

func (d dirTreeSeq) Len() (int, bool) {
	// For recursive directory walks, length is generally unknown
	// as it requires walking the entire tree to count
	return seq.LenUnknown, false
}

func (d dirTreeSeq) ToSlice() DirEntrySlice {
	var entries []fs.DirEntry

	err := filepath.WalkDir(d.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != d.dirName {
			entries = append(entries, entry)
		}
		return nil
	})

	if err != nil {
		return nil // errors silently dropped, matches pattern
	}

	return entries
}

func (d dirTreeSeq) Limit(n int) DirEntrySeq {
	return seq.LimitOf[fs.DirEntry](d, n)
}

func (d dirTreeSeq) Take(n int) (DirEntrySlice, DirEntrySeq) {
	var entries []fs.DirEntry
	taken := 0

	err := filepath.WalkDir(d.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != d.dirName {
			if taken < n {
				entries = append(entries, entry)
				taken++
			} else {
				// We've taken enough, stop walking
				return filepath.SkipAll
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return nil, seq.ErrorOf[fs.DirEntry](err)
	}

	if taken < n {
		// We took everything available
		return entries, seq.Empty[fs.DirEntry]()
	}

	// Create a sequence for the remaining entries by skipping what we've taken
	remaining := &skipDirTreeSeq{
		dirName: d.dirName,
		skip:    n,
	}
	return entries, remaining
}

func (d dirTreeSeq) TakeWhile(pred seq.Predicate[fs.DirEntry]) (DirEntrySlice, DirEntrySeq) {
	var entries []fs.DirEntry
	var stopped bool
	var stopIndex int

	err := filepath.WalkDir(d.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != d.dirName {
			if pred(entry) {
				entries = append(entries, entry)
			} else {
				stopped = true
				stopIndex = len(entries)
				return filepath.SkipAll
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return nil, seq.ErrorOf[fs.DirEntry](err)
	}

	if !stopped {
		return entries, seq.Empty[fs.DirEntry]()
	}

	// Create a sequence for the remaining entries
	remaining := &skipDirTreeSeq{
		dirName: d.dirName,
		skip:    stopIndex,
	}
	return entries, remaining
}

func (d dirTreeSeq) Skip(n int) DirEntrySeq {
	return &skipDirTreeSeq{
		dirName: d.dirName,
		skip:    n,
	}
}

func (d dirTreeSeq) Where(pred seq.Predicate[fs.DirEntry]) DirEntrySeq {
	return seq.WhereOf[fs.DirEntry](d, pred)
}

func (d dirTreeSeq) While(pred seq.Predicate[fs.DirEntry]) DirEntrySeq {
	return seq.WhileOf[fs.DirEntry](d, pred)
}

func (d dirTreeSeq) First() (opt.Opt[fs.DirEntry], DirEntrySeq) {
	var firstEntry fs.DirEntry
	var found bool

	err := filepath.WalkDir(d.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != d.dirName {
			firstEntry = entry
			found = true
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return opt.ErrorOf[fs.DirEntry](err), seq.ErrorOf[fs.DirEntry](err)
	}

	if !found {
		return opt.Empty[fs.DirEntry](), seq.Empty[fs.DirEntry]()
	}

	// Return the remaining entries by skipping the first one
	remaining := &skipDirTreeSeq{
		dirName: d.dirName,
		skip:    1,
	}
	return opt.Of(firstEntry), remaining
}

func (d dirTreeSeq) Map(shaper seq.FuncMap[fs.DirEntry, fs.DirEntry]) DirEntrySeq {
	return seq.MappingOf[fs.DirEntry](d, shaper)
}

// Error implements the contract for the seq.Error function.
func (d dirTreeSeq) Error() error {
	// Test if directory is readable
	_, err := os.Stat(d.dirName)
	return err
}

// skipDirTreeSeq is a helper for implementing Skip, Take, etc. on dirTreeSeq
type skipDirTreeSeq struct {
	dirName string
	skip    int
}

func (s *skipDirTreeSeq) ForEach(f seq.Func1[fs.DirEntry]) DirEntryOpt {
	skipped := 0
	err := filepath.WalkDir(s.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != s.dirName {
			if skipped < s.skip {
				skipped++
			} else {
				f(entry)
			}
		}
		return nil
	})

	if err != nil {
		return opt.ErrorOf[fs.DirEntry](err)
	}

	return DirEntryOpt{}
}

func (s *skipDirTreeSeq) ForEachIndex(f seq.Func2[int, fs.DirEntry]) DirEntryOpt {
	skipped := 0
	index := 0
	err := filepath.WalkDir(s.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != s.dirName {
			if skipped < s.skip {
				skipped++
			} else {
				f(index, entry)
				index++
			}
		}
		return nil
	})

	if err != nil {
		return opt.ErrorOf[fs.DirEntry](err)
	}

	return DirEntryOpt{}
}

func (s *skipDirTreeSeq) Len() (int, bool) {
	return seq.LenUnknown, false
}

func (s *skipDirTreeSeq) ToSlice() DirEntrySlice {
	var entries []fs.DirEntry
	skipped := 0

	err := filepath.WalkDir(s.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != s.dirName {
			if skipped < s.skip {
				skipped++
			} else {
				entries = append(entries, entry)
			}
		}
		return nil
	})

	if err != nil {
		return nil
	}

	return entries
}

func (s *skipDirTreeSeq) Limit(n int) DirEntrySeq {
	return seq.LimitOf[fs.DirEntry](s, n)
}

func (s *skipDirTreeSeq) Take(n int) (DirEntrySlice, DirEntrySeq) {
	var entries []fs.DirEntry
	skipped := 0
	taken := 0

	err := filepath.WalkDir(s.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != s.dirName {
			if skipped < s.skip {
				skipped++
			} else if taken < n {
				entries = append(entries, entry)
				taken++
			} else {
				return filepath.SkipAll
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return nil, seq.ErrorOf[fs.DirEntry](err)
	}

	if taken < n {
		return entries, seq.Empty[fs.DirEntry]()
	}

	remaining := &skipDirTreeSeq{
		dirName: s.dirName,
		skip:    s.skip + n,
	}
	return entries, remaining
}

func (s *skipDirTreeSeq) TakeWhile(pred seq.Predicate[fs.DirEntry]) (DirEntrySlice, DirEntrySeq) {
	var entries []fs.DirEntry
	skipped := 0
	var stopped bool
	stopCount := 0

	err := filepath.WalkDir(s.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != s.dirName {
			if skipped < s.skip {
				skipped++
			} else {
				if pred(entry) {
					entries = append(entries, entry)
					stopCount++
				} else {
					stopped = true
					return filepath.SkipAll
				}
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return nil, seq.ErrorOf[fs.DirEntry](err)
	}

	if !stopped {
		return entries, seq.Empty[fs.DirEntry]()
	}

	remaining := &skipDirTreeSeq{
		dirName: s.dirName,
		skip:    s.skip + stopCount,
	}
	return entries, remaining
}

func (s *skipDirTreeSeq) Skip(n int) DirEntrySeq {
	return &skipDirTreeSeq{
		dirName: s.dirName,
		skip:    s.skip + n,
	}
}

func (s *skipDirTreeSeq) Where(pred seq.Predicate[fs.DirEntry]) DirEntrySeq {
	return seq.WhereOf[fs.DirEntry](s, pred)
}

func (s *skipDirTreeSeq) While(pred seq.Predicate[fs.DirEntry]) DirEntrySeq {
	return seq.WhileOf[fs.DirEntry](s, pred)
}

func (s *skipDirTreeSeq) First() (opt.Opt[fs.DirEntry], DirEntrySeq) {
	skipped := 0
	var firstEntry fs.DirEntry
	var found bool

	err := filepath.WalkDir(s.dirName, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != s.dirName {
			if skipped < s.skip {
				skipped++
			} else {
				firstEntry = entry
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return opt.ErrorOf[fs.DirEntry](err), seq.ErrorOf[fs.DirEntry](err)
	}

	if !found {
		return opt.Empty[fs.DirEntry](), seq.Empty[fs.DirEntry]()
	}

	remaining := &skipDirTreeSeq{
		dirName: s.dirName,
		skip:    s.skip + 1,
	}
	return opt.Of(firstEntry), remaining
}

func (s *skipDirTreeSeq) Map(shaper seq.FuncMap[fs.DirEntry, fs.DirEntry]) DirEntrySeq {
	return seq.MappingOf[fs.DirEntry](s, shaper)
}

// Error implements the contract for the seq.Error function.
func (s *skipDirTreeSeq) Error() error {
	_, err := os.Stat(s.dirName)
	return err
}
