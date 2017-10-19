package model

import (
	"fmt"
	"github.com/golang/glog"

	"bufio"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type TfModel struct {
	Graph    *tf.Graph
	Labels   []string
	ModelDir string
}

func NewModel(mdir string) *TfModel {
	return &TfModel{
		ModelDir: mdir,
	}
}

func (m *TfModel) Init() error {
	if len(m.ModelDir) < 1 {
		glog.Errorf("modelDir is empty")
		return fmt.Errorf("modelDir is empty")
	}
	glog.V(2).Infof("begin to load model from: %v", m.ModelDir)

	modelfile, labelfile, err := m.modelFiles(m.ModelDir)
	if err != nil {
		err := fmt.Errorf("Failed to find model files in %v: %v", m.ModelDir, err)
		glog.Error(err.Error())
		return err
	}

	//1. load model
	glog.V(2).Infof("Begin to load model from %v", modelfile)
	model, err := ioutil.ReadFile(modelfile)
	if err != nil {
		err := fmt.Errorf("Failed ot read model file(%v): %v", modelfile, err)
		glog.Errorf(err.Error())
		return err
	}
	m.Graph = tf.NewGraph()
	if err := m.Graph.Import(model, ""); err != nil {
		glog.Error(err.Error())
		return err
	}

	//2. load labels
	if m.Labels, err = loadLabels(labelfile); err != nil {
		glog.Errorf("Failed to load labels(%v): %v", labelfile, err)
		return err
	}

	glog.V(2).Infof("Load %d labels from %v.", len(m.Labels), labelfile)
	return nil
}

/* return modelfile, labelfile, error */
func (m *TfModel) modelFiles(dir string) (string, string, error) {
	const URL = "https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip"

	modelfile := filepath.Join(dir, "tensorflow_inception_graph.pb")
	labelfile := filepath.Join(dir, "imagenet_comp_graph_label_strings.txt")
	zipfile := filepath.Join(dir, "inception5h.zip")

	if FilesExist(modelfile, labelfile) == nil {
		return modelfile, labelfile, nil
	}

	glog.Warningf("Did not find model in %v, will download from %v", dir, URL)
	if err := os.MkdirAll(dir, 0755); err != nil {
		glog.Errorf("Failed to create dir: %v", dir)
		return "", "", err
	}
	if err := download(URL, zipfile); err != nil {
		return "", "", fmt.Errorf("failed to download %v - %v", URL, err)
	}
	if err := unzip(dir, zipfile); err != nil {
		return "", "", fmt.Errorf("failed to extract contents from model archive: %v", err)
	}
	os.Remove(zipfile)
	return modelfile, labelfile, FilesExist(modelfile, labelfile)
}

func loadLabels(fname string) ([]string, error) {
	file, err := os.Open(fname)
	if err != nil {
		glog.Errorf(err.Error())
		return nil, err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	var labels []string
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		glog.Errorf("ERROR: failed to read %s: %v", fname, err)
		return nil, err
	}

	return labels, nil
}

func (m *TfModel) PredictTopkFile(fname string, k int) (*PredictResult, error) {
	glog.V(2).Infof("Begin to predict data from file %v", fname)
	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		glog.Errorf("failed to read image file %v: %v", fname, err)
		return nil, err
	}

	return m.PredictTopK(bytes, k)
}

func (m *TfModel) PredictTopK(bytes []byte, k int) (*PredictResult, error) {
	probabilities, err := m.PredictImage(bytes)
	if err != nil {
		glog.Errorf("Predict failed: %v", err)
		return nil, err
	}

	return m.getTopK(probabilities, k)
}

func (m *TfModel) PredictTopKTensor(tensor *tf.Tensor, k int) (*PredictResult, error) {
	probabilities, err := m.PredictTensor(tensor)
	if err != nil {
		glog.Errorf("Predict failed: %v", err)
		return nil, err
	}

	return m.getTopK(probabilities, k)
}

func (m *TfModel) getTopK(probabilities []float32, k int) (*PredictResult, error) {
	pairs := []*Pair{}
	for i, p := range probabilities {
		pair := &Pair{
			Index:  i,
			Weight: p,
		}
		pairs = append(pairs, pair)
	}

	sort.Sort(ByWeight(pairs))

	result := NewPredictResult()
	for i := 0; i < k; i++ {
		p := pairs[i]
		lw := NewLabelWeight(m.Labels[p.Index], p.Weight)
		result.Add(lw)
		if p.Weight < 0.0005 {
			break
		}
	}
	return result, nil
}

func (m *TfModel) PredictTensor(tensor *tf.Tensor) ([]float32, error) {
	result := []float32{}
	session, err := tf.NewSession(m.Graph, nil)
	if err != nil {
		glog.Errorf("Failed to create a new session to predict: %v", err)
		return result, err
	}
	defer timeTrack(time.Now(), "predict")
	defer session.Close()

	//3. execute the graph
	graph := m.Graph
	output, err := session.Run(
		map[tf.Output]*tf.Tensor{
			graph.Operation("input").Output(0): tensor,
		},
		[]tf.Output{
			graph.Operation("output").Output(0),
		},
		nil)
	if err != nil {
		glog.Errorf("Failed to run session to predict %v", err)
		return result, err
	}

	//4. get output
	probabilities := output[0].Value().([][]float32)[0]
	return probabilities, nil
}

func (m *TfModel) PredictImage(bytes []byte) ([]float32, error) {
	defer timeTrack(time.Now(), "predict.bytes.wallclock")
	result := []float32{}

	tensor, err := MakeTensorFromImage(bytes)
	if err != nil {
		glog.Errorf("Failed to construct tensor: %v", err)
		return result, err
	}

	//2. start the session
	session, err := tf.NewSession(m.Graph, nil)
	if err != nil {
		glog.Errorf("Failed to create a new session to predict: %v", err)
		return result, err
	}
	defer timeTrack(time.Now(), "predict")
	defer session.Close()

	//3. execute the graph
	graph := m.Graph
	output, err := session.Run(
		map[tf.Output]*tf.Tensor{
			graph.Operation("input").Output(0): tensor,
		},
		[]tf.Output{
			graph.Operation("output").Output(0),
		},
		nil)
	if err != nil {
		glog.Errorf("Failed to run session to predict %v", err)
		return result, err
	}

	//4. get output
	probabilities := output[0].Value().([][]float32)[0]
	return probabilities, nil
}

func (m *TfModel) PredictFile(fname string) ([]float32, error) {
	defer timeTrack(time.Now(), "predict.file.wallclock")
	result := []float32{}

	//1. prepare input for model
	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		glog.Errorf("Failed to load image from %v: %v", fname, err)
		return result, err
	}

	return m.PredictImage(bytes)
}

func MakeTensorFromFile(filename string) (*tf.Tensor, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		glog.Errorf("Failed to load image from %v: %v", filename, err)
		return nil, err
	}

	return MakeTensorFromImage(bytes)
}

func LoadImage(fname string) ([]byte, *tf.Tensor, error) {
	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		glog.Errorf("Failed to load image from %v: %v", fname, err)
		return []byte{}, nil, err
	}

	tensor, err := MakeTensorFromImage(bytes)
	if err != nil {
		glog.Errorf("Failed to construct tensor for file %v: %v", fname, err)
		return []byte{}, nil, err
	}

	return bytes, tensor, err
}

/*
makeTensorFromImage and constructGraphToNormlizeImage are copied from tensorflow.org.
*/
// Convert the image in filename to a Tensor suitable as input to the Inception model.
func MakeTensorFromImage(bytes []byte) (*tf.Tensor, error) {
	// DecodeJpeg uses a scalar String-valued tensor as input.
	tensor, err := tf.NewTensor(string(bytes))
	if err != nil {
		glog.Errorf("Failed to costruct tensor from bytes: %v", err)
		return nil, err
	}
	// Construct a graph to normalize the image
	graph, input, output, err := constructGraphToNormalizeImage()
	if err != nil {
		glog.Errorf("Failed to construct graph to normalize image: %v", err)
		return nil, err
	}
	// Execute that graph to normalize this one image
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		glog.Errorf("Failed to start session to normalize image: %v", err)
		return nil, err
	}
	defer session.Close()
	normalized, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		glog.Errorf("Failed to normalize image: %v", err)
		return nil, err
	}
	return normalized[0], nil
}

// The inception model takes as input the image described by a Tensor in a very
// specific normalized format (a particular image size, shape of the input tensor,
// normalized pixel values etc.).
//
// This function constructs a graph of TensorFlow operations which takes as
// input a JPEG-encoded string and returns a tensor suitable as input to the
// inception model.
func constructGraphToNormalizeImage() (graph *tf.Graph, input, output tf.Output, err error) {
	// Some constants specific to the pre-trained model at:
	// https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip
	//
	// - The model was trained after with images scaled to 224x224 pixels.
	// - The colors, represented as R, G, B in 1-byte each were converted to
	//   float using (value - Mean)/Scale.
	const (
		H, W  = 224, 224
		Mean  = float32(117)
		Scale = float32(1)
	)
	// - input is a String-Tensor, where the string the JPEG-encoded image.
	// - The inception model takes a 4D tensor of shape
	//   [BatchSize, Height, Width, Colors=3], where each pixel is
	//   represented as a triplet of floats
	// - Apply normalization on each pixel and use ExpandDims to make
	//   this single image be a "batch" of size 1 for ResizeBilinear.
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)
	output = op.Div(s,
		op.Sub(s,
			op.ResizeBilinear(s,
				op.ExpandDims(s,
					op.Cast(s,
						op.DecodeJpeg(s, input, op.DecodeJpegChannels(3)), tf.Float),
					op.Const(s.SubScope("make_batch"), int32(0))),
				op.Const(s.SubScope("size"), []int32{H, W})),
			op.Const(s.SubScope("mean"), Mean)),
		op.Const(s.SubScope("scale"), Scale))
	graph, err = s.Finalize()
	return graph, input, output, err
}
