package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
)

func main() {
	demoMode := flag.Bool("demo", false, "Run UI animation demo without touching files")
	flag.Parse()

	pterm.DefaultHeader.WithFullWidth().WithMargin(2).Println("FMove - Native GoPro Concatenator")

	var clips []VideoClip
	var dir string
	var err error

	if *demoMode {
		pterm.Info.Println("Running in UI DEMO MODE. No real files will be read or modified.")
		clips = []VideoClip{
			{Name: "GH010123.MP4", CreationTime: time.Now().Add(-2 * time.Hour), DurationUS: 420000000, Size: 4200 * 1024 * 1024},
			{Name: "GH010124.MP4", CreationTime: time.Now().Add(-1 * time.Hour), DurationUS: 360000000, Size: 3600 * 1024 * 1024},
			{Name: "GH010125.MP4", CreationTime: time.Now().Add(-10 * time.Minute), DurationUS: 150000000, Size: 1500 * 1024 * 1024},
		}
		dir = "/Demo/Workspace"
	} else {
		dir, err = os.Getwd()
		if err != nil {
			pterm.Fatal.Printf("Failed to get current directory: %v\n", err)
		}

		spinner, _ := pterm.DefaultSpinner.Start("Scanning directory for MP4 files natively...")
		clips, err = ScanDirectory(dir)
		if err != nil {
			spinner.Fail("Failed to scan directory")
			pterm.Fatal.Println(err)
		}

		if len(clips) == 0 {
			spinner.Warning("No MP4/MOV files found in the current directory.")
			return
		}

		spinner.Success(fmt.Sprintf("Found %d clips", len(clips)))
	}

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

	if *demoMode {
		runDemoAnimation(clips, totalDurationUS, deleteClips)
		return
	}

	// Start processing
	ProcessFiles(clips, outputFile, totalDurationUS, deleteClips)
}

func runDemoAnimation(clips []VideoClip, totalDurationUS int64, deleteClips bool) {
	p, _ := pterm.DefaultProgressbar.WithTotal(int(totalDurationUS / 1000000)).WithTitle("Concatenating clips...").Start()
	
	for _, clip := range clips {
		pterm.Info.Printf("Processing %s...\n", clip.Name)
		
		// Simulate processing time
		clipSeconds := int(clip.DurationUS / 1000000)
		for i := 0; i < clipSeconds; i++ {
			p.Add(1)
			time.Sleep(10 * time.Millisecond) // Fast animation
		}

		if deleteClips {
			pterm.Success.Printf("Processed and instantly freed %s!\n", clip.Name)
		} else {
			pterm.Success.Printf("Successfully processed %s\n", clip.Name)
		}
	}

	p.Stop()
	pterm.Success.Printf("\nDone! Saved to /Demo/Workspace/fmove_output.mp4\n")
}
