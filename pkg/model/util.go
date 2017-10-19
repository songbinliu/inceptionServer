package model

import (
	"archive/zip"
	"fmt"
	"github.com/golang/glog"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"bytes"
)

/* Pair is used to sort the prediction result. */
type Pair struct {
	Index  int
	Weight float32
}

type ByWeight []*Pair

func (a ByWeight) Len() int           { return len(a) }
func (a ByWeight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByWeight) Less(i, j int) bool { return a[i].Weight > a[j].Weight }

type labelWeight struct {
	Label string
	Weight float32
}

func NewLabelWeight(label string, weight float32) *labelWeight {
	return &labelWeight{
		Label: label,
		Weight: weight,
	}
}

type PredictResult struct {
	array []*labelWeight
}

func NewPredictResult() *PredictResult {
	return &PredictResult{}
}

func (r *PredictResult) Add(lw *labelWeight) {
	r.array = append(r.array, lw)
}

func (r *PredictResult) String() string {
	var buffer bytes.Buffer

	for i, lw := range r.array {
		w := fmt.Sprintf("\t[Top-%d] %2.1f%% likely ", i+1, lw.Weight*100)
		buffer.WriteString(w)
		buffer.WriteString(lw.Label)
		buffer.WriteString("\n")
	}

	return buffer.String()
}

func (r *PredictResult) GenTableString() string {
	var buf bytes.Buffer

	for _, lw := range r.array {
		buf.WriteString("<tr>")
		tmp := fmt.Sprintf("<td>%2.1f%% </td> ", lw.Weight*100)
		buf.WriteString(tmp)
		tmp = fmt.Sprintf("<td>%v</td>", lw.Label)
		buf.WriteString(tmp)
		buf.WriteString("</tr>")
	}

	return buf.String()
}


func timeTrack(start time.Time, name string) time.Duration{
	elapsed := time.Since(start)
	glog.V(2).Infof("%s took %s", name, elapsed)
	return elapsed
}

func filesExist(files ...string) error {
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			return fmt.Errorf("unable to stat %s: %v", f, err)
		}
	}
	return nil
}

func download(URL, filename string) error {
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

func unzip(dir, zipfile string) error {
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		src, err := f.Open()
		if err != nil {
			return err
		}
		glog.V(3).Infof("Extracting", f.Name)
		dst, err := os.OpenFile(filepath.Join(dir, f.Name), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
		dst.Close()
	}
	return nil
}
