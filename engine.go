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
	"sync"
	"time"
	"unsafe"
)

type VideoClip struct {
	Path         string
	CreationTime time.Time
	DurationUS   int64 // Microseconds
	Size         int64
	Name         string
	Width        int
	Height       int
	FPS          float64
}

func init() {
	C.init_ffmpeg()
}

func ScanDirectory(dir string, extFilter string, outputFile string) ([]VideoClip, error) {
	var clips []VideoClip
	var mu sync.Mutex
	var wg sync.WaitGroup

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if (ext == ".mp4" || ext == ".mov") && (extFilter == "" || ext == extFilter) {
			absPath, _ := filepath.Abs(filepath.Join(dir, entry.Name()))

			// Ignore our own potential output file
			if entry.Name() == outputFile {
				continue
			}

			wg.Add(1)
			go func(entry os.DirEntry, absPath string) {
				defer wg.Done()

				cPath := C.CString(absPath)
				meta := C.get_metadata(cPath)
				C.free(unsafe.Pointer(cPath))

				if meta != nil {
					info, _ := entry.Info()

					var creationTime time.Time
					if meta.creation_time != nil {
						timeStr := C.GoString(meta.creation_time)
						C.free(unsafe.Pointer(meta.creation_time))
						parsed, err := time.Parse(time.RFC3339Nano, timeStr)
						if err == nil {
							creationTime = parsed
						} else {
							creationTime = info.ModTime()
						}
					} else {
						creationTime = info.ModTime()
					}

					clip := VideoClip{
						Path:         absPath,
						CreationTime: creationTime,
						DurationUS:   int64(meta.duration),
						Size:         info.Size(),
						Name:         entry.Name(),
						Width:        int(meta.width),
						Height:       int(meta.height),
						FPS:          float64(meta.fps),
					}

					C.free(unsafe.Pointer(meta))

					mu.Lock()
					clips = append(clips, clip)
					mu.Unlock()
				}
			}(entry, absPath)
		}
	}

	wg.Wait()

	// Sort chronologically
	sort.Slice(clips, func(i, j int) bool {
		return clips[i].CreationTime.Before(clips[j].CreationTime)
	})

	return clips, nil
}

func ScanFiles(filePaths []string) ([]VideoClip, error) {
	clips := make([]VideoClip, len(filePaths))
	var wg sync.WaitGroup

	for i, path := range filePaths {
		wg.Add(1)
		go func(index int, absPath string) {
			defer wg.Done()

			cPath := C.CString(absPath)
			meta := C.get_metadata(cPath)
			C.free(unsafe.Pointer(cPath))

			if meta != nil {
				info, err := os.Stat(absPath)
				if err != nil {
					C.free(unsafe.Pointer(meta))
					return
				}

				var creationTime time.Time
				if meta.creation_time != nil {
					timeStr := C.GoString(meta.creation_time)
					C.free(unsafe.Pointer(meta.creation_time))
					parsed, err := time.Parse(time.RFC3339Nano, timeStr)
					if err == nil {
						creationTime = parsed
					} else {
						creationTime = info.ModTime()
					}
				} else {
					creationTime = info.ModTime()
				}

				clips[index] = VideoClip{
					Path:         absPath,
					CreationTime: creationTime,
					DurationUS:   int64(meta.duration),
					Size:         info.Size(),
					Name:         filepath.Base(absPath),
					Width:        int(meta.width),
					Height:       int(meta.height),
					FPS:          float64(meta.fps),
				}

				C.free(unsafe.Pointer(meta))
			}
		}(i, path)
	}

	wg.Wait()

	// Filter out any empty items (failed to scan)
	var validClips []VideoClip
	for _, clip := range clips {
		if clip.Path != "" {
			validClips = append(validClips, clip)
		}
	}

	return validClips, nil
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
