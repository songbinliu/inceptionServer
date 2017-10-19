#!/bin/bash

tag=beekman9527/inceptionserver
docker build -t $tag .
docker push $tag
