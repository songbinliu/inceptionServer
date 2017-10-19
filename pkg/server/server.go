package server

import (
	"fmt"
	"io"
	"time"
	"net/http"
	"html/template"
	"github.com/golang/glog"

	"inceptionServer/pkg/util"
	tfmodel "inceptionServer/pkg/model"
	"bytes"
	"os"
	"strings"
)


type InceptionServer struct {
	port int
	ip string
	host string

	model *tfmodel.TfModel
	imgDB *tfmodel.ImageDB

}

func NewInceptionServer(port int, m *tfmodel.TfModel) *InceptionServer {
	ip, err := util.ExternalIP()
	if err != nil {
		glog.Errorf("Failed to get server IP: %v", err)
		ip = "localhost"
	}

	host, err := os.Hostname()
	if err != nil {
		glog.Errorf("Failed to get hostname: %v", err)
		host = "localhost"
	}
	glog.V(2).Infof("Will server on %s:%d", ip, port)



	return &InceptionServer{
		port: port,
		ip: ip,
		host: host,
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

// handle pages "/", "/index.html", "index.htm"
func (s *InceptionServer) handleWelcome(w http.ResponseWriter, r *http.Request) {
	head, err := getHead("Welcome", "Introduction")
	if err != nil {
		glog.Errorf("Failed to handle welcome page.")
		io.WriteString(w, "Internal Error")
		return
	}

	body := `This is a web server, which can assign labels to images using tensorflow inception model. <br/>
	<a href="/img/random">Try it.</a>
	It will show a random image, and its labels.`

	foot := s.genPageFoot(r)

	io.WriteString(w, head + body + foot)
	return
}

func (s *InceptionServer) handlePredict(w http.ResponseWriter, r *http.Request) {
	glog.V(4).Infof("Begin to handle predict request: %v", r.URL.Path)
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

	//4. generate html
	foot := s.genPageFoot(r)
	//util.TimeTrack(begin, "Predict")
	io.WriteString(w, GetImgHtml(fname, bytes, htmlTable, foot, begin))
	return
}

func (s *InceptionServer) faviconHandler(w http.ResponseWriter, r *http.Request) {
	fpath := "/tmp/favicon.jpg"
	if err := tfmodel.FilesExist(fpath); err != nil {
		glog.Warningf("favicon file[%v] does not exist.", fpath)
		return
	}

	http.ServeFile(w, r, fpath)
	return
}

func (s *InceptionServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	glog.V(3).Infof("Begin to handle path: %v", path)

	if strings.EqualFold(path, "/favicon.ico") {
		s.faviconHandler(w, r)
		return
	}

	if strings.HasPrefix(path, "/img/") {
		s.handlePredict(w, r)
		return
	}

	s.handleWelcome(w, r)
	return
}

func (s *InceptionServer) genPageFoot (r *http.Request) string {
	tmp, err := template.New("foot").Parse(htmlFootTemplate)
	if err != nil {
		glog.Errorf("Failed to parse image template %v:%v", imageTemplate, err)
		return ""
	}

	var result bytes.Buffer

	data := make(map[string]interface{})
	data["HostName"] = s.host
	data["HostIP"] = s.ip
	data["ClientIP"] = getClientIP(r)
	data["OriginalClient"] = getOriginalClientInfo(r)

	if err := tmp.Execute(&result, data); err != nil {
		glog.Errorf("Faile to execute template: %v", err)
		return ""
	}

	return result.String()
}