package formatter

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestPrintGrid_ColumnFirst(t *testing.T) {
	tests := []struct {
		name      string
		items     []string
		margin    int
		termWidth int
		expected  string
	}{
		// General cases
		{
			name:      "Mixed names - terminal width 80",
			items:     []string{"file1.txt", "it's_long_name_document.pdf", "データ.csv", "image.png", "テスト", "another.file", " backup ", "絵文字😇.txt", "main.go", "util.go", "README.md", "LICENSE"},
			margin:    2,
			termWidth: 80,
			expected:  "file1.txt                      image.png     \" backup \"    util.go\n\"it's_long_name_document.pdf\"  テスト        絵文字😇.txt  README.md\nデータ.csv                     another.file  main.go       LICENSE\n",
		},
		{
			name:      "Short, medium, long - terminal width 60",
			items:     []string{"short", "medium_length", "a", "b", "c", "this_is_the_longest_one_for_sure", "d", "e", "f"},
			margin:    2,
			termWidth: 60,
			expected:  "short          a  c                                 d  f\nmedium_length  b  this_is_the_longest_one_for_sure  e\n",
		},
		{
			name:      "Empty list - terminal width 80",
			items:     []string{},
			margin:    2,
			termWidth: 80,
			expected:  "",
		},
		{
			name:      "Single item - terminal width 80",
			items:     []string{"lonely_file.txt"},
			margin:    2,
			termWidth: 80,
			expected:  "lonely_file.txt\n",
		},
		{
			name:      "Wide characters mix - terminal width 20",
			items:     []string{"a", "あ", "b", "い", "c", "う", "d", "え"},
			margin:    2,
			termWidth: 20,
			expected:  "a   b   c   d\nあ  い  う  え\n",
		},
		{
			name:      "Short and long",
			items:     []string{"a", "b", "ccccc", "ddddd"},
			margin:    2,
			termWidth: 10,
			expected:  "a  ccccc\nb  ddddd\n",
		},
		{
			name:      "With ANSI escape codes",
			items:     []string{"\033[31mhello\033[0m", "hoge"},
			margin:    2,
			termWidth: 13,
			expected:  "\033[31mhello\033[0m  hoge\n",
		},
		// Boundary conditions
		{
			name:      "Terminal width 1",
			items:     []string{"a", "bb"},
			margin:    2,
			termWidth: 1,
			expected:  "a\nbb\n",
		},
		{
			name:      "Margin 0 - fits",
			items:     []string{"aa", "bbb", "ccc"},
			margin:    0,
			termWidth: 10,
			expected:  "aa   ccc\nbbb\n",
		},
		{
			name:      "Margin equals terminal width",
			items:     []string{"a", "b"},
			margin:    10,
			termWidth: 10,
			expected:  "a\nb\n",
		},
		{
			name:      "Exactly fits - terminal width 36 (no trailing margin)",
			items:     []string{"short", "longer"},
			margin:    2,
			termWidth: 36,
			expected:  "short  longer\n",
		},
		{
			name:      "One item over fits - terminal width 30",
			items:     []string{"short", "longer_than_short_string"},
			margin:    2,
			termWidth: 30,
			expected:  "short\nlonger_than_short_string\n",
		},
		{
			name:      "Many short items - fits in one line",
			items:     []string{"a", "b", "c", "d", "e"},
			margin:    2,
			termWidth: 20,
			expected:  "a  b  c  d  e\n",
		},
		{
			name:      "Many short items - wraps to multiple lines",
			items:     []string{"a", "b", "c", "d", "e"},
			margin:    2,
			termWidth: 10,
			expected:  "a  c  e\nb  d\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printGrid(os.Stdout, tt.items, tt.margin, tt.termWidth, false)

			err := w.Close()
			if err != nil {
				t.Fatalf("Error closing pipe: %v", err)
			}

			var buf bytes.Buffer
			_, err = buf.ReadFrom(r)
			if err != nil {
				t.Fatalf("Error reading from pipe: %v", err)
			}
			os.Stdout = oldStdout

			actual := buf.String()
			if actual != tt.expected {
				t.Errorf("TestPrintGrid_ColumnFirst(%s) failed: expected\n%q\ngot\n%q", tt.name, tt.expected, actual)
			}
		})
	}
}

func TestPrintGrid_RowFirst(t *testing.T) {
	tests := []struct {
		name      string
		items     []string
		margin    int
		termWidth int
		expected  string
	}{
		// General cases
		{
			name:      "Mixed names - terminal width 80",
			items:     []string{"file1.txt", "it's_long_name_document.pdf", "データ.csv", "image.png", "テスト", "another.file", " backup ", "絵文字😇.txt", "main.go", "util.go", "README.md", "LICENSE"},
			margin:    2,
			termWidth: 80,
			expected:  "file1.txt  \"it's_long_name_document.pdf\"  データ.csv  image.png\nテスト     another.file                   \" backup \"  絵文字😇.txt\nmain.go    util.go                        README.md   LICENSE\n",
		},
		{
			name:      "Example 1-8 - terminal width 10",
			items:     []string{"1", "2", "3", "4", "5", "6", "7", "8"},
			margin:    2,
			termWidth: 10,
			expected:  "1  2  3\n4  5  6\n7  8\n",
		},
		{
			name:      "Short, medium, long - terminal width 50",
			items:     []string{"short", "medium_length", "a", "b", "c", "this_is_the_longest_one_for_sure", "d", "e", "f"},
			margin:    2,
			termWidth: 50,
			expected:  "short  medium_length\na      b\nc      this_is_the_longest_one_for_sure\nd      e\nf\n",
		},
		{
			name:      "Wide characters mix - terminal width 20",
			items:     []string{"a", "あ", "b", "い", "c", "う", "d", "え"},
			margin:    2,
			termWidth: 20,
			expected:  "a  あ  b  い\nc  う  d  え\n",
		},
		{
			name:      "Short and long",
			items:     []string{"a", "b", "ccccc", "ddddd"},
			margin:    2,
			termWidth: 10,
			expected:  "a\nb\nccccc\nddddd\n",
		},
		{
			name:      "Short and long 2",
			items:     []string{"a", "bbbbb", "c", "ddddd"},
			margin:    2,
			termWidth: 10,
			expected:  "a  bbbbb\nc  ddddd\n",
		},
		{
			name:      "Many short items - fits in one line",
			items:     []string{"a", "b", "c", "d", "e"},
			margin:    2,
			termWidth: 20,
			expected:  "a  b  c  d  e\n",
		},
		{
			name:      "Many short items - wraps to multiple lines",
			items:     []string{"a", "b", "c", "d", "e"},
			margin:    2,
			termWidth: 10,
			expected:  "a  b  c\nd  e\n",
		},
		{
			name:      "With ANSI escape codes",
			items:     []string{"\033[32mgreen\033[0m", "text"},
			margin:    3,
			termWidth: 15,
			expected:  "\033[32mgreen\033[0m   text\n",
		},
		// Boundary conditions
		{
			name:      "Exactly fits - terminal width 36 (no trailing margin)",
			items:     []string{"short", "longer"},
			margin:    2,
			termWidth: 36,
			expected:  "short  longer\n",
		},
		{
			name:      "One item over fits - terminal width 30",
			items:     []string{"short", "longer_than_short_string"},
			margin:    2,
			termWidth: 30,
			expected:  "short\nlonger_than_short_string\n",
		},
		{
			name:      "Margin zero - fits",
			items:     []string{"a", "bb", "ccc"},
			margin:    0,
			termWidth: 10,
			expected:  "a    bb\nccc\n",
		},
		{
			name:      "Margin equals terminal width",
			items:     []string{"a", "b"},
			margin:    10,
			termWidth: 10,
			expected:  "a\nb\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printGrid(os.Stdout, tt.items, tt.margin, tt.termWidth, true)

			err := w.Close()
			if err != nil {
				t.Fatalf("Error closing pipe: %v", err)
			}

			var buf bytes.Buffer
			_, err = buf.ReadFrom(r)
			if err != nil {
				t.Fatalf("Error reading from pipe: %v", err)
			}
			os.Stdout = oldStdout

			actual := buf.String()
			if actual != tt.expected {
				t.Errorf("TestPrintGrid_RowFirst(%s) failed: expected\n%q\ngot\n%q", tt.name, tt.expected, actual)
			}
		})
	}
}

func TestPrintGrid_NoTerminal(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected string
	}{
		{
			name:     "Single item",
			items:    []string{"test"},
			expected: "test\n",
		},
		{
			name:     "Multiple items",
			items:    []string{"item1", "item2", "item3"},
			expected: "item1\nitem2\nitem3\n",
		},
		{
			name:     "Empty list",
			items:    []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Simulate non-terminal by not checking term.IsTerminal
			FprintGrid(w, tt.items, 2, false) // byRow doesn't matter here

			err := w.Close()
			if err != nil {
				t.Fatalf("Error closing pipe: %v", err)
			}

			var buf bytes.Buffer
			_, err = buf.ReadFrom(r)
			if err != nil {
				t.Fatalf("Error reading from pipe: %v", err)
			}
			os.Stdout = oldStdout

			actual := buf.String()
			if actual != tt.expected {
				t.Errorf("TestPrintGrid_NoTerminal(%s) failed: expected\n%q\ngot\n%q", tt.name, tt.expected, actual)
			}
		})
	}
}

func TestPrintGrid_MinimumMargin(t *testing.T) {
	tests := []struct {
		name      string
		items     []string
		margin    int
		termWidth int
		expected  string
	}{
		{
			name:      "Margin less than minimum (column-first)",
			items:     []string{"a", "bb"},
			margin:    1,
			termWidth: 10,
			expected:  "a  bb\n",
		},
		{
			name:      "Margin less than minimum (row-first)",
			items:     []string{"a", "bb"},
			margin:    1,
			termWidth: 10,
			expected:  "a  bb\n",
		},
		{
			name:      "Margin equal to minimum (column-first)",
			items:     []string{"a", "bb"},
			margin:    2,
			termWidth: 10,
			expected:  "a  bb\n",
		},
		{
			name:      "Margin equal to minimum (row-first)",
			items:     []string{"a", "bb"},
			margin:    2,
			termWidth: 10,
			expected:  "a  bb\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printGrid(os.Stdout, tt.items, tt.margin, tt.termWidth, strings.Contains(tt.name, "(row-first)"))

			err := w.Close()
			if err != nil {
				t.Fatalf("Error closing pipe: %v", err)
			}

			var buf bytes.Buffer
			_, err = buf.ReadFrom(r)
			if err != nil {
				t.Fatalf("Error reading from pipe: %v", err)
			}
			os.Stdout = oldStdout

			actual := buf.String()
			if actual != tt.expected {
				t.Errorf("TestPrintGrid_MinimumMargin(%s) failed: expected\n%q\ngot\n%q", tt.name, tt.expected, actual)
			}
		})
	}
}
