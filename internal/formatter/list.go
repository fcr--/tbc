package formatter

import (
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strings"
	"tbc/internal/terabox"
	"tbc/internal/util"
	"time"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

const (
	quoteEscapeChars = "' "
	minMargin        = 2
	ansiEscapeRegExp = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
)

var ansiRegexp = regexp.MustCompile(ansiEscapeRegExp)

// AnsiStrip removes ANSI escape sequences from a string.
// This is useful for calculating the visual width of strings that might contain terminal control codes.
func AnsiStrip(str string) string {
	return ansiRegexp.ReplaceAllString(str, "")
}

// FprintGrid prints a slice of strings to the given io.Writer in a grid format.
// The grid layout adapts to the terminal width if the writer is a terminal.
//
// Parameters:
//
//	writer: The io.Writer to print to (e.g., os.Stdout).
//	items: The slice of strings to print.
//	margin: The number of spaces to use as a margin between columns.
//	byRow: A boolean indicating whether to print items row by row (true) or column by column (false).
func FprintGrid(writer io.Writer, items []string, margin int, byRow bool) {
	if len(items) == 0 {
		return
	}

	file, ok := writer.(*os.File)
	if !ok {
		// If the writer is not an os.File, print items line by line.
		for _, item := range items {
			fmt.Fprintln(writer, item)
		}
		return
	}

	stdoutFd := int(file.Fd())
	termWidth := -1

	if term.IsTerminal(stdoutFd) {
		width, _, err := term.GetSize(stdoutFd)
		if err != nil {
			// If getting terminal size fails, print items line by line.
			for _, item := range items {
				fmt.Fprintln(writer, item)
			}
			return
		}
		termWidth = width
	} else {
		// If not a terminal, print items line by line.
		for _, item := range items {
			fmt.Fprintln(writer, item)
		}
		return
	}

	// Print grid with determined terminal width.
	printGrid(writer, items, margin, termWidth, byRow)
}

// printGrid prints the items in a grid format with a specified terminal width.
// This function assumes that the output is a terminal and the terminal width has been determined.
func printGrid(writer io.Writer, items []string, margin int, termWidth int, byRow bool) {
	count := len(items)
	if count == 0 {
		return
	}

	margin = max(minMargin, margin)
	itemWidths := make([]int, count)
	for i := range itemWidths {
		itemWidths[i] = runewidth.StringWidth(AnsiStrip(items[i]))
		if strings.ContainsAny(items[i], quoteEscapeChars) {
			itemWidths[i] += 2 // Add padding for items containing single quote or space when quoted
		}
	}

	// inc determines the increment for the number of columns to efficiently find the best fit.
	inc := func(div int) int {
		if count <= div || div == 0 {
			return 1
		}
		b := float64(count)
		d := float64(div)
		return int(math.Ceil(b/(math.Ceil(b/d)-1))) - div
	}

	// getIndex calculates the index of the item in the slice based on the current column and row,
	// and whether the grid is printed by row or by column.
	getIndex := func(col, row, numCols, numRows int) int {
		if byRow {
			return row*numCols + col
		} else {
			return col*numRows + row
		}
	}

	bestNumCols := 1
	numRows := count
	colWidths := []int{0}

	// Iterate through possible numbers of columns to find the best fit within the terminal width.
	for currentNumCols := 1; currentNumCols <= count; currentNumCols += inc(currentNumCols) {
		currentNumRows := int(math.Ceil(float64(count) / float64(currentNumCols)))
		currentColWidths := make([]int, currentNumCols)
		currentTotalWidth := 0
		possible := true

		// Calculate the maximum width for each column.
		for col := range currentNumCols {
			maxWidthInCol := 0
			for row := range currentNumRows {
				index := getIndex(col, row, currentNumCols, currentNumRows)
				if index >= count {
					break
				}
				itemWidth := itemWidths[index]
				if itemWidth > maxWidthInCol {
					maxWidthInCol = itemWidth
				}
			}

			currentColWidths[col] = maxWidthInCol
			currentTotalWidth += maxWidthInCol + margin

			// If the current total width exceeds the terminal width, this layout is not possible.
			if currentTotalWidth > termWidth {
				possible = false
				break
			}
		}

		// If the current layout is possible, update the best layout found so far.
		if possible {
			bestNumCols = currentNumCols
			numRows = currentNumRows
			colWidths = currentColWidths
		} else if bestNumCols+2 < currentNumCols {
			// There may be a number of columns that fit best depending on how they are arranged,
			// so look for a little more.
			break
		}
	}

	// Print the grid using the best layout found.
	for row := range numRows {
		for col := range bestNumCols {
			index := getIndex(col, row, bestNumCols, numRows)
			if index < count {
				item := items[index]
				if strings.ContainsAny(item, quoteEscapeChars) {
					item = "\"" + item + "\"" // Quote items containing single quote or space
				}
				itemWidth := itemWidths[index]

				fmt.Fprint(writer, item)

				// Add padding to align items within columns and add margin between columns.
				if col < bestNumCols-1 && getIndex(col+1, row, bestNumCols, numRows) < count {
					padding := 0
					if col < len(colWidths) {
						padding = colWidths[col] - itemWidth
					}
					if padding < 0 {
						padding = 0
					}
					fmt.Fprint(writer, strings.Repeat(" ", padding+margin))
				}
			} else {
				break
			}
		}
		fmt.Fprintln(writer)
	}
}

type OptsLongFormat struct {
	HumanReadable bool
	Indicator     bool
	CreationTime  bool
	FileId        bool
}

func FprintLong(writer io.Writer, item *terabox.Item, opts *OptsLongFormat) {
	name := item.Name
	dir := "."
	if item.IsDir != 0 {
		dir = "d"
		if opts.Indicator {
			name = name + "/"
		}
	}

	share := "-"
	if item.Share != 0 {
		share = "*"
	}

	t := time.Unix(int64(item.Modified), 0)
	if opts.CreationTime {
		t = time.Unix(int64(item.Created), 0)
	}
	timeStr := t.Local().Format("2006-01-02 15:04:05")

	fileId := ""
	if opts.FileId {
		fileId = fmt.Sprintf(" %16d", item.FileId)
	}

	if opts.HumanReadable {
		size := util.ByteCountIEC(int64(item.Size))
		fmt.Fprintf(writer, "%s/%s%s %7s  %s %s\n", dir, share, fileId, size, timeStr, name)
	} else {
		fmt.Fprintf(writer, "%s/%s%s %11d  %s %s\n", dir, share, fileId, item.Size, timeStr, name)
	}
}
