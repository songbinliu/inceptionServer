#!/bin/bash
tag=beekman9527/inceptionserver:lit
docker build -f ./scripts/Dockerfilelit -t $tag .
docker push $tag
