# inceptionServer
A web server to assign labels to images using  tensorflow inception model.

<img width="293" alt="cat labels" src="https://user-images.githubusercontent.com/27221807/31775913-f41d29a6-b4b7-11e7-8457-b1a08a7f5304.png">

It will show a random image, and its labels. The image set is loaded from local filesystem `--imgdir`. New images can be added to this directory, and will be shown in the page.


# Run it
### Run it based on docker image
```bash
docker run -p 8080:9527 -d beekman9527/inceptionserver
```
Then It can be acccessed through http://localhost:8080/ .

Or load your own set of images for the server via:
```bash
myimgdir=/my/img/dir
docker run -p 8080:9527 -v $myimgdir:/tmp/imgs/ -d beekman9527/inceptionserver
```

Images can be found from [ImageNet](http://www.image-net.org).

# Build it
### Build local runnable binary (Mac OS X)

With the following commands, a executable binary will be built in `inceptionServer/_output/inceptions`.
```bash
cd inceptionServer
sh scripts/init.glide.sh
make build
sh scripts/run.sh
```

Then run this binary with following commands:
```bash
_output/inceptions --v=2 --logtostderr --modeldir=./model-data/inception/ --imgdir=./imgs/
```
A web server will be listening on port 9527. Access it via http://localhost:9527.

### Build a container image
```bash
sh scripts/build_img.sh
```
