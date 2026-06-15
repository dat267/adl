package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

// NamedEntity is a shared shape for circle, tag, and VA objects.
type NamedEntity struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// WorkInfo mirrors the full response from /api/workInfo/{id}.
type WorkInfo struct {
	// Core identity
	ID    int    `json:"id"`
	Title string `json:"title"`

	// Creator
	Circle NamedEntity `json:"circle"`

	// Voice actors and tags
	Vas  []NamedEntity `json:"vas"`
	Tags []NamedEntity `json:"tags"`

	// Artwork
	MainCoverURL string `json:"mainCoverUrl"`
	SamCoverURL  string `json:"samCover"`

	// Release metadata
	Release  string `json:"release"` // e.g. "2024-03-15"
	Language string `json:"language"`

	// Ratings and popularity
	Rate      float64 `json:"rate"`
	RateCount int     `json:"rateCount"`
	Rank      *int    `json:"rank"`

	// Pricing (in JPY)
	Price int `json:"price"`

	// Content flags
	NSFW bool `json:"nsfw"`

	// Work-level description
	Description string `json:"detail"` // may be HTML
}

// TrackItem mirrors one node from /api/tracks/{id}?v=2.
// Leaf nodes have MediaDownloadURL set; folder nodes have Children.
type TrackItem struct {
	Title            string      `json:"title"`
	Type             string      `json:"type"`             // "audio", "image", "text", "other"
	MediaDownloadURL string      `json:"mediaDownloadUrl"` // present on leaf nodes
	StreamURL        string      `json:"streamUrl"`        // HLS stream, if available
	Duration         float64     `json:"duration"`         // seconds
	Size             int64       `json:"size"`             // bytes
	Hash             string      `json:"hash"`             // MD5 or similar
	Children         []TrackItem `json:"children"`         // present on folder nodes
}

type TrackJob struct {
	FileName     string
	RelativePath string
	URL          string
}

var sanitizeReplacer = strings.NewReplacer(
	"\\", "-", "/", "-", ":", "-", "*", "-",
	"?", "-", "\"", "-", "<", "-", ">", "-", "|", "-",
)

func sanitize(s string) string {
	return strings.TrimSpace(sanitizeReplacer.Replace(s))
}

func extractID(input string) string {
	rjRegex := regexp.MustCompile(`(?i)RJ(\d+)`)
	if match := rjRegex.FindStringSubmatch(input); len(match) > 1 {
		return match[1]
	}
	if digitRegex.MatchString(input) {
		return input
	}
	return ""
}

func extractTracks(items []TrackItem, currentPath string, exclude *regexp.Regexp, results []TrackJob) []TrackJob {
	for _, item := range items {
		title := sanitize(item.Title)
		if exclude != nil && exclude.MatchString(title) {
			continue
		}
		if len(item.Children) > 0 {
			results = extractTracks(item.Children, filepath.Join(currentPath, title), exclude, results)
		} else if item.MediaDownloadURL != "" {
			results = append(results, TrackJob{
				FileName:     title,
				RelativePath: currentPath,
				URL:          item.MediaDownloadURL,
			})
		}
	}
	return results
}

func downloadFile(url, destPath string) error {
	if _, err := os.Stat(destPath); err == nil {
		fmt.Printf("- Skipping (Already exists): %s\n", filepath.Base(destPath))
		return nil
	}

	targetDir := filepath.Dir(destPath)
	tempDir := filepath.Join(targetDir, ".downloading")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return err
	}

	tempPath := filepath.Join(tempDir, filepath.Base(destPath)+".tmp")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}

	out, err := os.Create(tempPath)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(tempPath)
		return err
	}
	out.Close()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		os.Remove(tempPath)
		return err
	}

	if err := os.Rename(tempPath, destPath); err != nil {
		os.Remove(tempPath)
		return err
	}

	fmt.Printf("- Downloaded: %s\n", filepath.Base(destPath))
	return nil
}

func fetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func processWork(id, baseDownloadDir string, exclude *regexp.Regexp, preferFlac bool, sem chan struct{}) {
	var info WorkInfo
	if err := fetchJSON(fmt.Sprintf("https://api.asmr-200.com/api/workInfo/%s", id), &info); err != nil {
		fmt.Printf("! Error fetching info for RJ%s: %v\n", id, err)
		return
	}

	var rawTracks []TrackItem
	resp, err := http.Get(fmt.Sprintf("https://api.asmr-200.com/api/tracks/%s?v=2", id))
	if err != nil {
		fmt.Printf("! Error fetching tracks for RJ%s: %v\n", id, err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyBytes, &rawTracks); err != nil {
		var singleTrack TrackItem
		if err := json.Unmarshal(bodyBytes, &singleTrack); err != nil {
			fmt.Printf("! Error parsing tracks for RJ%s\n", id)
			return
		}
		rawTracks = []TrackItem{singleTrack}
	}

	tracks := extractTracks(rawTracks, "", exclude, nil)
	circleName := sanitize(info.Circle.Name)
	if circleName == "" {
		circleName = "Unknown Circle"
	}
	workFolderName := sanitize(fmt.Sprintf("[RJ%s] %s", id, info.Title))
	rootFolder := filepath.Join(baseDownloadDir, circleName, workFolderName)

	fmt.Printf("Processing RJ%s: %s\nTarget: %s\n", id, info.Title, rootFolder)

	var wg sync.WaitGroup
	uniqueDirs := make(map[string]bool)

	for _, track := range tracks {
		targetDir := filepath.Join(rootFolder, track.RelativePath)
		destPath := filepath.Join(targetDir, track.FileName)
		uniqueDirs[targetDir] = true

		if preferFlac && strings.HasSuffix(strings.ToLower(track.FileName), ".wav") {
			flacPath := destPath[:len(destPath)-4] + ".flac"
			if _, err := os.Stat(flacPath); err == nil {
				fmt.Printf("- Skipping (Local FLAC exists): %s\n", track.FileName)
				continue
			}
		}

		wg.Add(1)
		go func(t TrackJob, p string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := downloadFile(t.URL, p); err != nil {
				fmt.Printf("! Failed %s: %v\n", t.FileName, err)
			}
		}(track, destPath)
	}

	wg.Wait()

	for dir := range uniqueDirs {
		os.RemoveAll(filepath.Join(dir, ".downloading"))
	}
}

func main() {
	excludePattern := flag.String("exclude", "", "Regex pattern to exclude tracks by title")
	customDir := flag.String("dir", "", "Custom base download directory")
	preferFlac := flag.Bool("prefer-flac", false, "Skip downloading WAV if matching FLAC exists locally")
	concurrency := flag.Int("concurrency", 1, "Number of concurrent file downloads")
	flag.Parse()

	var excludeRegex *regexp.Regexp
	if *excludePattern != "" {
		excludeRegex = regexp.MustCompile(*excludePattern)
	}

	baseDir := *customDir
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not find home directory")
			os.Exit(1)
		}
		baseDir = filepath.Join(home, "Audio", "ASMR")
	}

	var ids []string
	for _, arg := range flag.Args() {
		if id := extractID(arg); id != "" {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		fmt.Printf("%s %s\nUsage: %s [options] <ID or URL>...\n",
			filepath.Base(os.Args[0]), version, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(1)
	}

	sem := make(chan struct{}, *concurrency)
	for _, id := range ids {
		processWork(id, baseDir, excludeRegex, *preferFlac, sem)
	}

	fmt.Println("All tasks complete.")
}
