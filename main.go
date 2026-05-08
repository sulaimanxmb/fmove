package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
)

func main() {
	pterm.DefaultHeader.WithFullWidth().WithMargin(2).Println("FMove - Native GoPro Concatenator")

	dir, err := os.Getwd()
	if err != nil {
		pterm.Fatal.Printf("Failed to get current directory: %v\n", err)
	}

	spinner, _ := pterm.DefaultSpinner.Start("Scanning directory for MP4 files natively...")
	clips, err := ScanDirectory(dir)
	if err != nil {
		spinner.Fail("Failed to scan directory")
		pterm.Fatal.Println(err)
	}

	if len(clips) == 0 {
		spinner.Warning("No MP4/MOV files found in the current directory.")
		return
	}

	spinner.Success(fmt.Sprintf("Found %d clips", len(clips)))

	// Build Table Data
	tableData := pterm.TableData{
		{"#", "Filename", "Recording Time", "Duration", "Size"},
	}

	var totalDurationUS int64
	var totalSize int64

	for i, clip := range clips {
		durStr := time.Duration(clip.DurationUS * 1000).String()
		sizeStr := fmt.Sprintf("%.2f MB", float64(clip.Size)/(1024*1024))

		tableData = append(tableData, []string{
			fmt.Sprintf("%d", i+1),
			clip.Name,
			clip.CreationTime.Format("15:04:05"),
			durStr,
			sizeStr,
		})

		totalDurationUS += clip.DurationUS
		totalSize += clip.Size
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	pterm.Println()

	expectedSize := fmt.Sprintf("%.2f GB", float64(totalSize)/(1024*1024*1024))
	pterm.Info.Printf("Total Expected Duration: %s | Total Expected Size: %s\n\n", time.Duration(totalDurationUS*1000).String(), expectedSize)

	// User confirmation
	var result string
	fmt.Print("Ready to concatenate and dynamically free space. Type 'y' to continue: ")
	fmt.Scanln(&result)
	if result != "y" && result != "Y" {
		pterm.Warning.Println("Operation cancelled by user.")
		return
	}

	outputFile := filepath.Join(dir, "fmove_output.mp4")

	// Start processing
	ProcessFiles(clips, outputFile, totalDurationUS)
}
