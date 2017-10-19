package model

import (
	"fmt"
	"math/rand"
	"github.com/golang/glog"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

type ImageDB struct {
	images map[string]*tf.Tensor
	rawImages map[string][]byte
	index map[int]string
}

func NewImageDB () *ImageDB {
	images := make(map[string]*tf.Tensor)
	rawImages := make(map[string][]byte)
	index := make(map[int]string)

	return &ImageDB{
		images: images,
		rawImages: rawImages,
		index: index,
	}
}

func (db *ImageDB) Add(fname string, tensor *tf.Tensor, bytes []byte) {
	i := db.Size()
	db.images[fname] = tensor
	db.rawImages[fname] = bytes
	db.index[i] = fname
}

func (db *ImageDB) Load(fname string) error {
	bytes, tensor, err := LoadImage(fname)
	if err != nil {
		glog.Errorf("failed to generate tensor from file %v: %v", fname, err)
		return err
	}

	db.Add(fname, tensor, bytes)
	return nil
}

func (db *ImageDB) Print() {
	fmt.Printf("Number of Images: %d\n", len(db.images))
	for fname, bytes := range db.rawImages {
		fmt.Printf("\t%v : %d\n", fname, len(bytes))
	}
}

func (db *ImageDB) Size() int {
	return len(db.rawImages)
}

func (db *ImageDB) Get(fname string) (*tf.Tensor, []byte, error) {
	bytes := []byte{}
	tensor, ok := db.images[fname]
	if !ok {
		return nil, bytes, fmt.Errorf("%s not exists", fname)
	}

	bytes, ok = db.rawImages[fname]
	if !ok {
		return nil, bytes, fmt.Errorf("%s not exist", fname)
	}

	return tensor, bytes, nil
}

func (db *ImageDB) GetTensor(fname string) (*tf.Tensor, error) {
	tensor, ok := db.images[fname]
	if !ok {
		return nil, fmt.Errorf("%s not exists", fname)
	}
	return tensor, nil
}

func (db *ImageDB) GetRawImage(fname string) ([]byte, error) {
	bytes, ok := db.rawImages[fname]
	if !ok {
		glog.Errorf("%s not exists", fname)
		return bytes, fmt.Errorf("Not exist")
	}
	return bytes, nil
}

func (db *ImageDB) GetRandomImage() (string, error) {
	if db.Size() < 1 {
		glog.Errorf("ImageDB is empty.")
		return "", fmt.Errorf("Empty.")
	}

	id := rand.Int31n(int32(db.Size()))
	return db.index[int(id)], nil
}