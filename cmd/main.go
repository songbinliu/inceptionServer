package main

import (
	"fmt"
	"flag"
	"math/rand"
	"time"
	"github.com/golang/glog"
	"path/filepath"
	"io/ioutil"
	"strings"
	"runtime"

	tfmodel "inceptionServer/pkg/model"
	iserver "inceptionServer/pkg/server"

)

var (
	modeldir string
	imgfile  string
	imgdir string
	port int
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func setFlags() error {
	flag.StringVar(&modeldir, "modeldir", "./model-data/inception/", "model directory")
	flag.StringVar(&imgfile, "imgfile", "", "path to the image file, for example ./imgs/cat.jpg")
	flag.StringVar(&imgdir, "imgdir", "/tmp/imgs/", "path to the image files")
	flag.IntVar(&port, "port", 9527, "port to listen on")

	flag.Parse()
	if modeldir == "" {
		fmt.Println("modeldir must be provided.")
		flag.Usage()
		return fmt.Errorf("wrong parameter")
	}

	return nil
}

func loadImages(dir string) (*tfmodel.ImageDB, error) {
	imgDB := tfmodel.NewImageDB()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		glog.Errorf("Failed to readDir %v: %v", dir, err)
		return nil, fmt.Errorf("Failed to load data: %v", err)
	}

	for _, file := range files {
		fname := file.Name()
		if !strings.HasSuffix(fname, "jpg") {
			continue
		}

		fname = filepath.Join(dir, file.Name())
		if err := imgDB.Load(fname); err != nil {
			glog.Errorf("failed to generate tensor from file %v: %v", fname, err)
			continue
		}
	}

	if imgDB.Size() < 1 {
		return nil, fmt.Errorf("No jpg image in dir %v", dir)
	}

	return imgDB, nil
}

func testImageDB(db *tfmodel.ImageDB, model *tfmodel.TfModel) {
	fname, err := db.GetRandomImage()
	if err != nil {
		glog.Errorf("Failed to fecth a fname.")
		return
	}

	tensor, err := db.GetTensor(fname)
	if err != nil {
		glog.Errorf("Failed to get tensor from ImageDB: %v", err)
		return
	}

	fmt.Printf("fname:%v\n", fname)
	result, err := model.PredictTopKTensor(tensor, 5)
	if err != nil {
		glog.Errorf("Failed to predict %v: %v", fname, err)
		return
	}

	fmt.Println(result.String())
}

func testFile(imgfile string, model *tfmodel.TfModel) {
	result, err := model.PredictTopkFile(imgfile, 5)
	if err != nil {
		glog.Errorf("Failed to predict %v: %v", imgfile, err)
		return
	}

	fmt.Println(result.String())
}


func main() {
	if err := setFlags(); err != nil {
		fmt.Println("Wrong parameters")
		return
	}

	//1. load the model
	model := tfmodel.NewModel(modeldir)
	if err := model.Init(); err != nil {
		glog.Errorf("Failed to load model %v: %v", modeldir, err)
		return
	}
	glog.V(2).Infof("Load model(%v) successfully.", modeldir)

	if len(imgfile) > 0 {
		testFile(imgfile, model)
	}

	//2. load the images, and transform it
	images, err := loadImages(imgdir)
	if err != nil {
		glog.Errorf("Failed to load images from dir %v: %v", imgdir, err)
		return
	}
	testImageDB(images, model)

	//3. construct the server
	server := iserver.NewInceptionServer(port, model)
	server.SetImages(images)
	server.Print()
	server.Run()

	glog.V(2).Infof("hello")
	return
}

