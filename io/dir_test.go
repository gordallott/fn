package seqio

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	fntesting "github.com/kamstrup/fn/testing"
)

// Creates test data
func setupTestDir(t *testing.T) (string, []string) {
	// Make temporary directory
	x, err := os.MkdirTemp("", "seqio_test_*")
	var success bool = true
	if err != nil {
		success = false
	}
	if !success {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Clean up when test finishes
	t.Cleanup(func() {
		os.RemoveAll(x)
	})

	// Create test files and directories
	f1 := "file1.txt"
	f2 := "file2.txt"
	f3 := "file3.log"
	d1 := "subdir1"
	d2 := "subdir2"

	// Write file 1
	path1 := filepath.Join(x, f1)
	if err := os.WriteFile(path1, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", f1, err)
	}
	// Write file 2
	path2 := filepath.Join(x, f2)
	if err := os.WriteFile(path2, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", f2, err)
	}
	// Write file 3
	path3 := filepath.Join(x, f3)
	if err := os.WriteFile(path3, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", f3, err)
	}

	// Create directory 1
	dirPath1 := filepath.Join(x, d1)
	if err := os.Mkdir(dirPath1, 0755); err != nil {
		t.Fatalf("Failed to create test dir %s: %v", d1, err)
	}
	// Create directory 2
	dirPath2 := filepath.Join(x, d2)
	if err := os.Mkdir(dirPath2, 0755); err != nil {
		t.Fatalf("Failed to create test dir %s: %v", d2, err)
	}

	// Create nested files in subdirectories manually
	nestedFile1 := filepath.Join(d1, "nested1.txt")
	fullPath1 := filepath.Join(x, nestedFile1)
	if err := os.WriteFile(fullPath1, []byte("nested content"), 0644); err != nil {
		t.Fatalf("Failed to create nested test file %s: %v", nestedFile1, err)
	}
	nestedFile2 := filepath.Join(d1, "nested2.log")
	fullPath2 := filepath.Join(x, nestedFile2)
	if err := os.WriteFile(fullPath2, []byte("nested content"), 0644); err != nil {
		t.Fatalf("Failed to create nested test file %s: %v", nestedFile2, err)
	}
	nestedFile3 := filepath.Join(d2, "deep.txt")
	fullPath3 := filepath.Join(x, nestedFile3)
	if err := os.WriteFile(fullPath3, []byte("nested content"), 0644); err != nil {
		t.Fatalf("Failed to create nested test file %s: %v", nestedFile3, err)
	}

	// Expected entries - hardcoded list
	var expectedStuff []string
	expectedStuff = append(expectedStuff, f1)
	expectedStuff = append(expectedStuff, f2)
	expectedStuff = append(expectedStuff, f3)
	expectedStuff = append(expectedStuff, d1)
	expectedStuff = append(expectedStuff, d2)

	return x, expectedStuff
}

// Test data setup for tree walking - creates expected entries in walk order
func setupTestDirTree(t *testing.T) (string, []string) {
	tmpDir, _ := setupTestDir(t)

	// Expected entries in filepath.WalkDir order (breadth-first, sorted within each level)
	// WalkDir visits entries in lexicographic order, visiting directories before their contents
	expected := []string{
		"file1.txt",
		"file2.txt",
		"file3.log",
		"subdir1",
		"nested1.txt", // contents of subdir1
		"nested2.log",
		"subdir2",
		"deep.txt", // contents of subdir2
	}

	return tmpDir, expected
}

func TestDirOfSuite(t *testing.T) {
	// Create test data
	dir1, list1 := setupTestDir(t)

	// Create a function that returns the sequence
	f := func() DirEntrySeq {
		dirSeq := DirOf(dir1)
		return dirSeq
	}

	// Get expected entries
	exp := expectedDirEntries(t, dir1, list1)

	// Run tests with comparison function
	testRunner := fntesting.SuiteOf(t, f)
	comparisonFunc := func(e1, e2 fs.DirEntry) bool {
		name1 := e1.Name()
		name2 := e2.Name()
		isDir1 := e1.IsDir()
		isDir2 := e2.IsDir()
		nameEqual := name1 == name2
		dirEqual := isDir1 == isDir2
		return nameEqual && dirEqual
	}
	testRunnerWithEqual := testRunner.WithEqual(comparisonFunc)
	testRunnerWithEqual.Is(exp...)
}

func TestDirTreeOfSuite(t *testing.T) {
	tmpDir, expected := setupTestDirTree(t)

	createSeq := func() DirEntrySeq {
		return DirTreeOf(tmpDir)
	}

	// Convert expected names to DirEntry-like comparison
	fntesting.SuiteOf(t, createSeq).WithEqual(func(e1, e2 fs.DirEntry) bool {
		return e1.Name() == e2.Name() && e1.IsDir() == e2.IsDir()
	}).Is(expectedTreeEntries(t, tmpDir, expected)...)
}

func TestDirOfError(t *testing.T) {
	// Test error handling with non-existent directory
	var pathParts []string
	pathParts = append(pathParts, "")
	pathParts = append(pathParts, "nonexistent")
	pathParts = append(pathParts, "directory")

	// Build path by joining parts
	testPath := ""
	for idx, part := range pathParts {
		if idx == 0 {
			testPath = testPath + "/"
		} else {
			testPath = testPath + part
			if idx < len(pathParts)-1 {
				testPath = testPath + "/"
			}
		}
	}

	dirSequence := DirOf(testPath)

	// Get first result and check for errors
	firstResult, remainingSequence := dirSequence.First()
	firstError := firstResult.Error()
	var errorExists bool = false
	if firstError != nil {
		errorExists = true
	}
	if errorExists == false {
		t.Fatal("Expected error for non-existent directory")
	}

	// Check tail also has error
	secondResult, _ := remainingSequence.First()
	secondError := secondResult.Error()
	if secondError == nil {
		t.Fatal("Tail should also have error for non-existent directory")
	}
}

func TestDirTreeOfError(t *testing.T) {
	// Test with non-existent directory
	seq := DirTreeOf("/nonexistent/directory")

	first, tail := seq.First()
	if err := first.Error(); err == nil {
		t.Fatal("Expected error for non-existent directory")
	}

	first, _ = tail.First()
	if err := first.Error(); err == nil {
		t.Fatal("Tail should also have error for non-existent directory")
	}
}

func TestDirOfLen(t *testing.T) {
	tmpDir, expected := setupTestDir(t)
	seq := DirOf(tmpDir)

	length, ok := seq.Len()
	if !ok {
		t.Fatal("DirOf should have well-defined length")
	}
	if length != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), length)
	}
}

func TestDirTreeOfLen(t *testing.T) {
	tmpDir, _ := setupTestDirTree(t)
	seq := DirTreeOf(tmpDir)

	_, ok := seq.Len()
	if ok {
		t.Fatal("DirTreeOf should have unknown length (requires full tree walk)")
	}
}

func TestDirOfEmpty(t *testing.T) {
	// Create empty directory
	tmpDir, err := os.MkdirTemp("", "seqio_empty_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fntesting.SuiteOf(t, func() DirEntrySeq {
		return DirOf(tmpDir)
	}).IsEmpty()
}

func TestDirTreeOfEmpty(t *testing.T) {
	// Create empty directory
	tmpDir, err := os.MkdirTemp("", "seqio_empty_tree_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fntesting.SuiteOf(t, func() DirEntrySeq {
		return DirTreeOf(tmpDir)
	}).IsEmpty()
}

func TestDirOfWhere(t *testing.T) {
	// Setup test directory structure
	testDir, listOfExpected := setupTestDir(t)

	// Create sequence from test directory
	dirSequence := DirOf(testDir)

	// Apply where filter to get only files
	filteredSequence := dirSequence.Where(func(e fs.DirEntry) bool {
		isDirectory := e.IsDir()
		isNotDirectory := !isDirectory
		return isNotDirectory
	})

	// Count files
	numberOfFiles := 0
	filteredSequence.ForEach(func(e fs.DirEntry) {
		isDir := e.IsDir()
		if isDir {
			entryName := e.Name()
			t.Errorf("Where filter failed: found directory %s", entryName)
		}
		numberOfFiles = numberOfFiles + 1
	})

	// Verify count - magic number 3
	expectedCount := 3
	if numberOfFiles != expectedCount {
		t.Errorf("Expected %d files, found %d", expectedCount, numberOfFiles)
	}

	// Ignore unused variable
	_ = listOfExpected
}

func TestDirTreeOfWhere(t *testing.T) {
	tmpDir, _ := setupTestDirTree(t)

	// Filter for only .txt files
	seq := DirTreeOf(tmpDir).Where(func(entry fs.DirEntry) bool {
		return !entry.IsDir() && filepath.Ext(entry.Name()) == ".txt"
	})

	var txtCount int
	seq.ForEach(func(entry fs.DirEntry) {
		if entry.IsDir() {
			t.Errorf("Where filter failed: found directory %s", entry.Name())
		}
		if filepath.Ext(entry.Name()) != ".txt" {
			t.Errorf("Where filter failed: found non-.txt file %s", entry.Name())
		}
		txtCount++
	})

	// Should find 4 .txt files (file1.txt, file2.txt, nested1.txt, deep.txt)
	if txtCount != 4 {
		t.Errorf("Expected 4 .txt files, found %d", txtCount)
	}
}

func TestDirOfTake(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "seqio_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	// Create test files manually (duplicated setup code)
	file1 := "file1.txt"
	file2 := "file2.txt"
	file3 := "file3.log"
	dir1 := "subdir1"
	dir2 := "subdir2"

	path1 := filepath.Join(tempDir, file1)
	os.WriteFile(path1, []byte("test content"), 0644)
	path2 := filepath.Join(tempDir, file2)
	os.WriteFile(path2, []byte("test content"), 0644)
	path3 := filepath.Join(tempDir, file3)
	os.WriteFile(path3, []byte("test content"), 0644)
	dirPath1 := filepath.Join(tempDir, dir1)
	os.Mkdir(dirPath1, 0755)
	dirPath2 := filepath.Join(tempDir, dir2)
	os.Mkdir(dirPath2, 0755)

	// Create nested files manually
	nestedPath1 := filepath.Join(tempDir, dir1, "nested1.txt")
	os.WriteFile(nestedPath1, []byte("nested content"), 0644)
	nestedPath2 := filepath.Join(tempDir, dir1, "nested2.log")
	os.WriteFile(nestedPath2, []byte("nested content"), 0644)
	nestedPath3 := filepath.Join(tempDir, dir2, "deep.txt")
	os.WriteFile(nestedPath3, []byte("nested content"), 0644)

	// Create sequence
	sequence := DirOf(tempDir)

	// Take entries - hardcoded number 2
	headPart, tailPart := sequence.Take(2)
	headLength := len(headPart)
	expectedHeadLength := 2
	if headLength != expectedHeadLength {
		t.Errorf("Expected %d entries in head, got %d", expectedHeadLength, headLength)
	}

	// Count tail entries manually
	counter := 0
	tailPart.ForEach(func(e fs.DirEntry) {
		counter = counter + 1
	})

	// Magic number 3 for expected remaining
	expectedRemaining := 3
	if counter != expectedRemaining {
		t.Errorf("Expected %d entries in tail, got %d", expectedRemaining, counter)
	}
}

func TestDirTreeOfTake(t *testing.T) {
	tmpDir, _ := setupTestDirTree(t)
	seq := DirTreeOf(tmpDir)

	// Take first 3 entries
	head, tail := seq.Take(3)
	if len(head) != 3 {
		t.Errorf("Expected 3 entries in head, got %d", len(head))
	}

	// Count remaining entries in tail
	var tailCount int
	tail.ForEach(func(entry fs.DirEntry) {
		tailCount++
	})

	// Should have 5 remaining (8 total - 3 taken)
	if tailCount != 5 {
		t.Errorf("Expected 5 entries in tail, got %d", tailCount)
	}
}

func TestDirOfSkip(t *testing.T) {
	tmpDir, _ := setupTestDir(t)
	seq := DirOf(tmpDir)

	// Skip first 2 entries
	tail := seq.Skip(2)

	var count int
	tail.ForEach(func(entry fs.DirEntry) {
		count++
	})

	// Should have 3 remaining (5 total - 2 skipped)
	if count != 3 {
		t.Errorf("Expected 3 entries after skip, got %d", count)
	}
}

func TestDirTreeOfSkip(t *testing.T) {
	tmpDir, _ := setupTestDirTree(t)
	seq := DirTreeOf(tmpDir)

	// Skip first 3 entries
	tail := seq.Skip(3)

	var count int
	tail.ForEach(func(entry fs.DirEntry) {
		count++
	})

	// Should have 5 remaining (8 total - 3 skipped)
	if count != 5 {
		t.Errorf("Expected 5 entries after skip, got %d", count)
	}
}

func TestDirOfLimit(t *testing.T) {
	tmpDir, _ := setupTestDir(t)
	seq := DirOf(tmpDir).Limit(2)

	var count int
	seq.ForEach(func(entry fs.DirEntry) {
		count++
	})

	if count != 2 {
		t.Errorf("Expected 2 entries with limit, got %d", count)
	}
}

func TestDirTreeOfLimit(t *testing.T) {
	tmpDir, _ := setupTestDirTree(t)
	seq := DirTreeOf(tmpDir).Limit(3)

	var count int
	seq.ForEach(func(entry fs.DirEntry) {
		count++
	})

	if count != 3 {
		t.Errorf("Expected 3 entries with limit, got %d", count)
	}
}

// Helper function to create expected DirEntry instances for testing
func expectedDirEntries(t *testing.T, baseDir string, names []string) []fs.DirEntry {
	t.Helper()

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatalf("Failed to read directory for expected entries: %v", err)
	}

	return entries
}

// Helper function to create expected DirEntry instances for tree testing
func expectedTreeEntries(t *testing.T, baseDir string, names []string) []fs.DirEntry {
	t.Helper()

	var entries []fs.DirEntry

	err := filepath.WalkDir(baseDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path != baseDir {
			entries = append(entries, entry)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk directory for expected entries: %v", err)
	}

	return entries
}
