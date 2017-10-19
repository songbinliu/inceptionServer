# inceptionServer
A web server to assign labels to images using  tensorflow inception model.

<img width="293" alt="cat labels" src="https://user-images.githubusercontent.com/27221807/31775913-f41d29a6-b4b7-11e7-8457-b1a08a7f5304.png">

It will show a random image, and its labels. The image set is loaded from local filesystem `--imgdir`. New images can be added to this directory, and will be shown in the page.


# Run it
### Run it based on docker image
```
docker run -p 8080:9527 -d beekman9527/inceptionserver
```
Then It can be acccessed through http://localhost:8080/ .


