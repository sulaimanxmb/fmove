package main

/*
#cgo pkg-config: libavformat libavcodec libavutil
#include "ffmpeg.h"
*/
import "C"

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unsafe"
)

type VideoClip struct {
	Path         string
	CreationTime time.Time
	DurationUS   int64 // Microseconds
	Size         int64
	Name         string
}

func init() {
	C.init_ffmpeg()
}

func ScanDirectory(dir string) ([]VideoClip, error) {
	var clips []VideoClip

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".mp4" || ext == ".mov" {
			absPath, _ := filepath.Abs(filepath.Join(dir, entry.Name()))

			// Ignore our own potential output file
			if entry.Name() == "fmove_output.mp4" {
				continue
			}

			cPath := C.CString(absPath)

			// Get Duration
			duration := int64(C.get_duration(cPath))

			// Get Creation Time
			cTimeStr := C.get_creation_time(cPath)
			var creationTime time.Time
			if cTimeStr != nil {
				timeStr := C.GoString(cTimeStr)
				C.free(unsafe.Pointer(cTimeStr))
				// FFmpeg creation_time usually format: 2024-03-22T10:15:30.000000Z
				parsed, err := time.Parse(time.RFC3339Nano, timeStr)
				if err == nil {
					creationTime = parsed
				} else {
					// Fallback
					info, _ := entry.Info()
					creationTime = info.ModTime()
				}
			} else {
				info, _ := entry.Info()
				creationTime = info.ModTime()
			}

			info, _ := entry.Info()

			clips = append(clips, VideoClip{
				Path:         absPath,
				CreationTime: creationTime,
				DurationUS:   duration,
				Size:         info.Size(),
				Name:         entry.Name(),
			})

			C.free(unsafe.Pointer(cPath))
		}
	}

	// Sort chronologically
	sort.Slice(clips, func(i, j int) bool {
		return clips[i].CreationTime.Before(clips[j].CreationTime)
	})

	return clips, nil
}

// Global variable for progress callback
var currentProgressCallback func(int64)

//export progressCallback
func progressCallback(current_time_us C.int64_t) {
	if currentProgressCallback != nil {
		currentProgressCallback(int64(current_time_us))
	}
}

// Helper to convert C function pointer for callback
func setCallback(cb func(int64)) {
	currentProgressCallback = cb
}
