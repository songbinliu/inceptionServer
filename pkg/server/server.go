package server

import (
	"fmt"
	"io"
	"time"
	"net/http"
	"github.com/golang/glog"

	"inceptionServer/pkg/util"
	tfmodel "inceptionServer/pkg/model"
)


type InceptionServer struct {
	port int

	model *tfmodel.TfModel
	imgDB *tfmodel.ImageDB
}

func NewInceptionServer(port int, m *tfmodel.TfModel) *InceptionServer {
	return &InceptionServer{
		port: port,
		model: m,
	}
}

func (s *InceptionServer) Print() {
	fmt.Printf("Number of labels: %d\n", len(s.model.Labels))
	s.imgDB.Print()
}

func (s *InceptionServer) SetImages(imgs *tfmodel.ImageDB) {
	s.imgDB = imgs
}

func (s *InceptionServer) Run() {
	server := http.Server {
		Addr: fmt.Sprintf(":%d", s.port),
		Handler: s,
	}

	glog.V(1).Infof("HTTP Server listens on: %s", server.Addr)
	panic(server.ListenAndServe())
}

func (s *InceptionServer) doPredict(fname string) (string, error) {
	tensor, err := s.imgDB.GetTensor(fname)
	if err != nil {
		glog.Errorf("Failed to get tensor for %v: %v", err, fname)
		return "", err
	}

	result, err := s.model.PredictTopKTensor(tensor, 5)
	if err != nil {
		glog.Errorf("Failed to predict image %v: %v", fname, err)
		return "", err
	}

	return result.GenTableString(), nil
}

func (s *InceptionServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//io.WriteString(w, "hello world!")
	//io.WriteString(w, GetSimpleHtml())
	begin := time.Now()

	//1. get a random image
	fname, err := s.imgDB.GetRandomImage()
	if err != nil {
		glog.Errorf("Failed to get an image: %v", err)
		io.WriteString(w, "Internal Error")
		return
	}

	//2. predict the labels for the image
	htmlTable, err := s.doPredict(fname)
	if err != nil {
		io.WriteString(w, "Internal Error")
		return
	}

	//3. write result
	bytes, err := s.imgDB.GetRawImage(fname)
	if err != nil {
		io.WriteString(w, "Internal Error")
		return
	}

	util.TimeTrack(begin, "Predict")
	io.WriteString(w, GetImgHtml(fname, bytes, htmlTable, begin))
	return
}