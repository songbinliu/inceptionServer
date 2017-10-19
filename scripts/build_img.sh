#!/bin/bash

tag=beekman9527/inceptionServer
docker build -t $tag .
docker push $tag
