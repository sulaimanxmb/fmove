package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

var AppVersion = "dev"

func main() {
	demoMode := flag.Bool("animation", false, "Run UI animation demo without touching files")
	outFlag := flag.String("o", "fmove_output.mp4", "Output filename")
	combineFlag := flag.String("c", "", "Combine specific files (comma separated) or all files in a directory (e.g., -c clip1.mp4,clip2.mp4 or -c .)")
	forceFlag := flag.Bool("f", false, "Force space recovery mode without interactive prompt")
	extFlag := flag.String("ext", "", "Filter by specific extension (e.g., .mov)")
	silentFlag := flag.Bool("s", false, "Silent mode: hide UI and banners")
	versionFlag := flag.Bool("v", false, "Print version information")
	flag.BoolVar(versionFlag, "version", false, "Print version information")

	flag.Usage = func() {
		fmt.Printf("FMove - Native Space Recovery & Concatenator %s\n\n", AppVersion)
		fmt.Printf("Usage: fmove [options]\n\n")
		fmt.Printf("Examples:\n")
		fmt.Printf("  fmove -c .                        (Combine all videos in current directory)\n")
		fmt.Printf("  fmove -c clip1.mp4,clip2.mp4      (Combine specific files)\n")
		fmt.Printf("  fmove -c /path/to/vids -o out.mp4 (Combine directory into specific output)\n\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Printf("FMove Native Concatenator %s\n", AppVersion)
		os.Exit(0)
	}

	if !*silentFlag {
		printBanner()
	}

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
		if *combineFlag == "" {
			pterm.Error.Println("Please provide input files or a directory using the -c flag. Run 'fmove --help' for details.")
			os.Exit(1)
		}

		var spinner *pterm.SpinnerPrinter
		if !*silentFlag {
			spinner, _ = pterm.DefaultSpinner.Start("Scanning files natively...")
		}

		if strings.Contains(*combineFlag, ",") {
			// Comma separated files
			files := strings.Split(*combineFlag, ",")
			for i := range files {
				files[i] = strings.TrimSpace(files[i])
				files[i], _ = filepath.Abs(files[i])
			}
			clips, err = ScanFiles(files)
			if len(files) > 0 {
				dir = filepath.Dir(files[0])
			}
		} else {
			// Directory or Single file
			info, errStat := os.Stat(*combineFlag)
			if errStat != nil {
				if spinner != nil {
					spinner.Fail("Failed to read input path")
				}
				pterm.Error.Println(errStat)
				os.Exit(1)
			}
			if info.IsDir() {
				dir, _ = filepath.Abs(*combineFlag)
				clips, err = ScanDirectory(dir, *extFlag, *outFlag)
			} else {
				absPath, _ := filepath.Abs(*combineFlag)
				dir = filepath.Dir(absPath)
				clips, err = ScanFiles([]string{absPath})
			}
		}

		if err != nil {
			if spinner != nil {
				spinner.Fail("Failed to scan files")
			}
			pterm.Error.Println(err)
			os.Exit(1)
		}

		if len(clips) == 0 {
			if spinner != nil {
				spinner.Warning("No matching video files found.")
			}
			return
		}

		if spinner != nil {
			spinner.Success(fmt.Sprintf("Found %d clips", len(clips)))
		}
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

	if !*silentFlag {
		pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		pterm.Println()

		expectedSize := fmt.Sprintf("%.2f GB", float64(totalSize)/(1024*1024*1024))
		pterm.Info.Printf("Total Expected Duration: %s | Total Expected Size: %s\n\n", time.Duration(totalDurationUS*1000).String(), expectedSize)
	}

	if !checkMismatches(clips, *silentFlag, *forceFlag) {
		return
	}

	continueOp, deleteClips := getUserConfirmation(*silentFlag, *forceFlag)
	if !continueOp {
		return
	}

	outputFile := filepath.Join(dir, *outFlag)

	if *demoMode {
		runDemoAnimation(clips, totalDurationUS, deleteClips)
		return
	}

	// Start processing
	ProcessFiles(clips, outputFile, totalDurationUS, deleteClips, *silentFlag)
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

func printBanner() {
	bannerLines := []string{
		`  ______  __       __   ______   ____    ____  ______`,
		` |   ___||  \     /  | /  __  \  \   \  /   / |   ___|  ` + AppVersion,
		` |  |__  |   \   /   ||  |  |  |  \   \/   /  |  |__`,
		` |   __| |    \_/    ||  |  |  |   \      /   |   __|`,
		` |  |    |  |\   /|  ||  '--'  |    \    /    |  |____`,
		` |__|    |__| \_/ |__| \______/      \__/     |_______|  A Native Space Recovery & Concatenator`,
	}

	colors := []pterm.Color{
		pterm.FgLightMagenta,
		pterm.FgLightCyan,
		pterm.FgLightGreen,
		pterm.FgLightYellow,
		pterm.FgLightRed,
		pterm.FgLightBlue,
	}

	fmt.Println()
	for i, line := range bannerLines {
		pterm.NewStyle(colors[i%len(colors)], pterm.Bold).Println(line)
	}

	fmt.Println()
	fmt.Println(pterm.Blue("                                                                                       by Sulaiman\n"))
}

func checkMismatches(clips []VideoClip, silentFlag, forceFlag bool) bool {
	if len(clips) == 0 {
		return true
	}

	firstWidth := clips[0].Width
	firstHeight := clips[0].Height
	firstFPS := clips[0].FPS

	mismatch := false
	for _, clip := range clips {
		if clip.Width != firstWidth || clip.Height != firstHeight {
			mismatch = true
		}
		if math.Abs(clip.FPS-firstFPS) > 0.1 {
			mismatch = true
		}
	}

	if mismatch && !silentFlag {
		pterm.Warning.Println("WARNING: Resolution or Framerate mismatch detected between clips!")
		pterm.Warning.Println("Concatenating files with different dimensions or framerates may result in playback corruption or audio desync.")
		if !forceFlag {
			confirm, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Do you want to proceed anyway?").Show()
			if !confirm {
				pterm.Warning.Println("Operation cancelled by user.")
				return false
			}
		} else {
			pterm.Warning.Println("Force flag provided. Proceeding despite mismatches.")
		}
	}
	return true
}

func getUserConfirmation(silentFlag, forceFlag bool) (bool, bool) {
	if forceFlag {
		if !silentFlag {
			pterm.Warning.Println("Force flag provided. Automatically entering Space Recovery Mode.")
		}
		return true, true
	}

	options := []string{
		"1. Normal Concatenation (Keep original clips)",
		"2. Space Recovery Mode (Dynamically delete original clips)",
		"Cancel",
	}

	selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).WithDefaultText("Select operation mode").Show()

	if selectedOption == "Cancel" || selectedOption == "" {
		pterm.Warning.Println("Operation cancelled by user.")
		return false, false
	}

	if selectedOption == options[1] {
		pterm.Warning.Println("WARNING: You have selected Space Recovery Mode.")
		pterm.Warning.Println("Original clips will be permanently deleted from your drive immediately after their packets are processed!")

		confirm, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Are you absolutely sure you want to delete the source files?").Show()
		if !confirm {
			pterm.Warning.Println("Operation cancelled by user.")
			return false, false
		}
		return true, true
	}

	return true, false
}
