package processor

import (
	"github.com/boyter/cs/file"
	"runtime"
	"sync"

	"io/ioutil"
	"os"

	"sync/atomic"
)

type FileReaderWorker struct {
	input      chan *file.File
	output     chan *fileJob
	fileCount  int64 // Count of the number of files that have been read
	InstanceId int
	SearchPDF  bool
}

func NewFileReaderWorker(input chan *file.File, output chan *fileJob) FileReaderWorker {
	return FileReaderWorker{
		input:     input,
		output:    output,
		fileCount: 0,
	}
}

func (f *FileReaderWorker) GetFileCount() int64 {
	return atomic.LoadInt64(&f.fileCount)
}

// This is responsible for spinning up all of the jobs
// that read files from disk into memory
func (f *FileReaderWorker) Start() {
	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			for res := range f.input {
				extension := file.GetExtension(res.Filename)

				switch extension {
				case "pdf":
					if SearchPDF {
						f.processPdf(res)
					}
				default:
					f.processUnknown(res)
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	close(f.output)
}

// For PDF if we are running in HTTP or TUI mode we really want to have
// a cache because the conversion can be expensive
var __pdfCache = map[string]string{}

func (f *FileReaderWorker) processPdf(res *file.File) {

	c, ok := __pdfCache[res.Location]
	if ok {
		atomic.AddInt64(&f.fileCount, 1)
		f.output <- &fileJob{
			Filename:       res.Filename,
			Extension:      "",
			Location:       res.Location,
			Content:        []byte(c),
			Bytes:          0,
			Score:          0,
			MatchLocations: map[string][][]int{},
		}
		return
	}

	content, err := convertPDFTextPdf2Txt(res.Location)
	if err != nil {
		content, err = convertPDFText(res.Location)
	}

	if err != nil {
		return
	}

	// Cache the result for PDF
	__pdfCache[res.Location] = content

	atomic.AddInt64(&f.fileCount, 1)
	f.output <- &fileJob{
		Filename:       res.Filename,
		Extension:      "",
		Location:       res.Location,
		Content:        []byte(content),
		Bytes:          0,
		Score:          0,
		MatchLocations: map[string][][]int{},
	}
}

func (f *FileReaderWorker) processUnknown(res *file.File) {
	fi, err := os.Stat(res.Location)
	if err != nil {
		return
	}

	var content []byte
	var s int64 = 10000000  // 10 MB in decimal counting

	// Only read up to ~10MB of a file because anything beyond that is probably pointless
	if fi.Size() < s {
		content, err = ioutil.ReadFile(res.Location)
	} else {
		f, err := os.Open(res.Location)
		if err != nil {
			return
		}
		defer f.Close()

		byteSlice := make([]byte, s)
		_, err = f.Read(byteSlice)
		if err != nil {
			return
		}

		content = byteSlice
	}

	if err == nil {
		atomic.AddInt64(&f.fileCount, 1)
		f.output <- &fileJob{
			Filename:       res.Filename,
			Extension:      "",
			Location:       res.Location,
			Content:        content,
			Bytes:          0,
			Score:          0,
			MatchLocations: map[string][][]int{},
		}
	}
}
