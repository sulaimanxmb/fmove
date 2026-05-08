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

	// User confirmation via interactive select
	options := []string{
		"1. Normal Concatenation (Keep original clips)",
		"2. Space Recovery Mode (Dynamically delete original clips)",
		"Cancel",
	}

	selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).WithDefaultText("Select operation mode").Show()

	if selectedOption == "Cancel" || selectedOption == "" {
		pterm.Warning.Println("Operation cancelled by user.")
		return
	}

	deleteClips := false
	if selectedOption == options[1] {
		pterm.Warning.Println("WARNING: You have selected Space Recovery Mode.")
		pterm.Warning.Println("Original clips will be permanently deleted from your drive immediately after their packets are processed!")
		
		confirm, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Are you absolutely sure you want to delete the source files?").Show()
		if !confirm {
			pterm.Warning.Println("Operation cancelled by user.")
			return
		}
		deleteClips = true
	}

	outputFile := filepath.Join(dir, "fmove_output.mp4")

	// Start processing
	ProcessFiles(clips, outputFile, totalDurationUS, deleteClips)
}
