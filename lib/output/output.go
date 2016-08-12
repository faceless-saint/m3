package output

import (
	"fmt"
	"github.com/cavaliercoder/grab"
	"os"
	"strings"
	"time"
)

type DownloadTracker struct {
	Name      string
	Channel   <-chan *grab.Response
	Responses []*grab.Response
	Interval  time.Duration
	Count     int
	Total     int
}

type DirectDownloadTracker struct {
	Name     string
	Response *grab.Response
	Interval time.Duration
}

// Log records the download progress without any ANSI terminal handling
// or other fancy features, ideal for output to files or DOS terminals.
func (this *DownloadTracker) Log() {
	if this.Count == 0 {
		fmt.Printf("%d %s found. Nothing to download.\n", this.Total, this.Name)
		return
	}

	// Print header
	fmt.Printf("Downloading %d %s... (%d found locally)\n", this.Count, this.Name, this.Total-this.Count)

	// Set up timer for download tracker
	t := time.NewTicker(this.Interval * time.Millisecond)
	defer t.Stop()

	completed := 0
	responses := []*grab.Response{}
	for completed < this.Count {
		select {
		case resp := <-this.Channel:
			if resp != nil {
				// Add new responses to the list as downloads start
				responses = append(responses, resp)
			}
		case <-t.C:
			for i, resp := range responses {
				// Log completed response
				if resp != nil && resp.IsComplete() {
					completed++
					if resp.Error != nil {
						fmt.Fprintf(os.Stderr, "\t%s - err: %v\n", strings.Replace(resp.Filename, this.Name+"/", "", 1), resp.Error)
					} else {
						fmt.Printf("\t%s\n", strings.Replace(resp.Filename, this.Name+"/", "", 1))
					}
					this.Responses = append(this.Responses, resp)
					responses[i] = nil
				}
			}
		}
	}
}

/*
func trackDownloadStatus(
        ch <-chan *grab.Response,
        timer time.Duration,
        name string,
        max, total, concurrency int) {
    t := time.NewTicker(timer * time.Millisecond)
    defer t.Stop()

    completed := 0
    inProgress := 0
    lastProgress := 0
    errors := make([]*string, 0)
    responses := make([]*grab.Response, 0)
    fmt.Print("\033[K\n")
    for completed < total {
        select {
        case resp := <-ch:
            // Add responses to the list as they arrive
            if resp != nil {
                responses = append(responses, resp)
            }
        case <-t.C:
            // At regular intervals, check and report progress
			if inProgress > 0 {
				fmt.Printf("\033[%dA\033[K", inProgress)
			}
            fmt.Printf("\033[1A    Downloading %s... (%d/%d)\033[K\n", name, completed, total)

			lastProgress = inProgress
			inProgress = 0
			for i, resp := range responses {
				if resp != nil {
					if resp.IsComplete() {
						// File download has completed
						if resp.Error != nil {
							errors = append(errors, &resp.Filename)
						}
						completed++
						responses[i] = nil
						if total-completed < concurrency {
							inProgress++
							fmt.Print("\033[K\n")
						}
					} else {
						// File download is still in progress
						inProgress++
						fmt.Printf("        (%3d%%) %s\033[K\n",
							int(100*resp.Progress()), resp.Filename)
					}
				}
			}
			if inProgress < lastProgress {
				fmt.Printf("%s\033[%dA\033[K",
					strings.Repeat("\n", lastProgress-inProgress+1),
					lastProgress-inProgress+1)
			}
        }
    }
    if lastProgress > 0 {
        fmt.Printf("\033[%dA", inProgress)
    }
    // Print a summary of the mod download results
    fmt.Printf("\033[1A\033[K%s\t %3d | %3d | %3d\n",
        name,
        total-len(errors),
		max-total,
		len(errors))
	for _, file := range errors {
		fmt.Fprintf(os.Stderr, "    error: %v\n", *file)
	}
}
func pretty_download_status(
	s *spec.Spec,
	respch <-chan *grab.Response,
	timer time.Duration,
	total, length, concurrency int) {
	fmt.Printf("Configuring mod directory...\n\n")
	// Set up the timer for download progress updates
	t := time.NewTicker(timer * time.Millisecond)
	defer t.Stop()

	// Track mod download progress
	completed := 0
	inProgress := 0
	lastProgress := 0
	errors := make([]*string, 0)
	responses := make([]*grab.Response, 0)
	for completed < total {
		select {
		case resp := <-respch:
			// Add responses to the list as they arrive
			if resp != nil {
				responses = append(responses, resp)
			}

		case <-t.C:
			// At regular intervals, check and report progress
			if inProgress > 0 {
				fmt.Printf("\033[%dA\033[K", inProgress)
			}

			// Draw progress bar
			print_progress_bar(completed, total, length)

			lastProgress = inProgress
			inProgress = 0
			for i, resp := range responses {
				if resp != nil {
					if resp.IsComplete() {
						// File download has completed
						if resp.Error != nil {
							errors = append(errors, &resp.Filename)
						}
						completed++
						responses[i] = nil
						if total-completed < concurrency {
							inProgress++
							fmt.Print("\033[K\n")
						}
					} else {
						// File download is still in progress
						inProgress++
						fmt.Printf("  (%3d%%) %s\033[K\n",
							int(100*resp.Progress()), resp.Filename)
					}
				}
			}
			if inProgress < lastProgress {
				fmt.Printf("%s\033[%dA\033[K",
					strings.Repeat("\n", lastProgress-inProgress+1),
					lastProgress-inProgress+1)
			}
		}
	}
	// Clear the progress bar
	fmt.Printf("\033[%dA\033[KConfiguring mod directory... Done.",
		inProgress+2)

	// Print a summary of the mod download results
	fmt.Printf("\t(%d local, %d remote, %d failed)\n",
		len(s.Mods.Items)-total,
		total-len(errors),
		len(errors))
	for _, file := range errors {
		fmt.Fprintf(os.Stderr, "\t%s\n", *file)
	}
}
func verbose_download_status(
	s *spec.Spec,
	respch <-chan *grab.Response,
	timer time.Duration,
	total int,
	very_verbose bool) {
	fmt.Printf("Configuring mod directory...\n")
	// Set up the timer for download progress updates
	t := time.NewTicker(timer * time.Millisecond)
	defer t.Stop()

	// Track mod download progress
	completed := make([]*string, 0)
	errors := make([]*string, 0)
	responses := make([]*grab.Response, 0)
	for len(completed) < total {
		select {
		case resp := <-respch:
			// Add responses to the list as they arrive
			if resp != nil {
				responses = append(responses, resp)
				fmt.Print("\n")
			}

		case <-t.C:
			// At regular intervals, check and report progress
			fmt.Printf("\033[%dA\033[K", len(responses)+1)
			fmt.Printf("Configuring mod directory... %d/%d\033[K\n", len(completed), total)

			for _, str := range completed {
				fmt.Print(*str)
			}
			for i, resp := range responses {
				if resp != nil {
					sum := hex.EncodeToString(resp.Request.Checksum)
					if sum != "" {
						sum = sum[:9]
					} else {
						sum = "no digest"
					}
					output_str := fmt.Sprintf(
						"  (%3d%%) %s - %s\t[%s]\t%s\033[K\n",
						int(100*resp.Progress()),
						netutil.ByteCountToString(resp.BytesTransferred()),
						netutil.ByteCountToString(resp.Size), sum,
						strings.Replace(resp.Filename, "mods/", "", 1))
					if resp.IsComplete() {
						// File download has completed
						if resp.Error != nil {
							errors = append(errors, &resp.Filename)
						}
						completed = append(completed, &output_str)
						responses[i] = nil
					}
					fmt.Print(output_str)
				}
			}
		}
	}
	// Print a summary of the mod download results
	fmt.Printf("\t(%d local, %d remote, %d failed)\n",
		len(s.Mods.Items)-total,
		total-len(errors),
		len(errors))
	for _, file := range errors {
		fmt.Fprintf(os.Stderr, "\t%s\n", *file)
	}
}
func print_progress_bar(progress, max, len int) {
	fmt.Printf("\033[1A[%s%s] %d/%d\033[K\n",
		strings.Repeat("=", int(len*progress/max)),
		strings.Repeat(" ", int(len-len*progress/max)),
		progress, max)
}*/
