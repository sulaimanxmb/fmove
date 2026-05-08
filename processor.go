package main

/*
#cgo pkg-config: libavformat libavcodec libavutil
#include "ffmpeg.h"
*/
import "C"
import (
	"os"
	"unsafe"

	"github.com/pterm/pterm"
)

var pbar *pterm.ProgressbarPrinter

func ProcessFiles(clips []VideoClip, outputFilename string, totalDurationUS int64) {
	if len(clips) == 0 {
		return
	}

	cFirstInput := C.CString(clips[0].Path)
	cOutput := C.CString(outputFilename)
	defer C.free(unsafe.Pointer(cFirstInput))
	defer C.free(unsafe.Pointer(cOutput))

	state := C.init_output(cFirstInput, cOutput)
	if state == nil {
		pterm.Fatal.Println("Failed to initialize output FFmpeg context")
	}
	defer C.finalize_output(state)

	p, _ := pterm.DefaultProgressbar.WithTotal(int(totalDurationUS / 1000000)).WithTitle("Concatenating clips...").Start()
	pbar = p

	var accumulatedTimeUS int64 = 0

	setCallback(func(currentUS int64) {
		// Update progress bar
		seconds := int((accumulatedTimeUS + currentUS) / 1000000)
		if seconds > pbar.Current {
			pbar.Add(seconds - pbar.Current)
		}
	})

	for _, clip := range clips {
		pterm.Info.Printf("Processing %s...\n", clip.Name)

		cPath := C.CString(clip.Path)
		ret := C.append_file(state, cPath)
		C.free(unsafe.Pointer(cPath))

		if ret < 0 {
			pterm.Error.Printf("Failed to append %s\n", clip.Name)
			continue
		}

		accumulatedTimeUS += clip.DurationUS

		// The core feature: Delete immediately after processing
		err := os.Remove(clip.Path)
		if err != nil {
			pterm.Warning.Printf("Processed %s, but failed to delete: %v\n", clip.Name, err)
		} else {
			pterm.Success.Printf("Processed and instantly freed %s!\n", clip.Name)
		}
	}

	pbar.Stop()
	pterm.Success.Printf("\nDone! Saved to %s\n", outputFilename)
}
