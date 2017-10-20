#!/bin/sh 

serverbin=/bin/inceptions

logv=3
alsologtostderr=true
modeldir="/tmp/model-data/inception"
imgdir="/tmp/imgs/"


for i in "$@"
do
case $i in
    -v=*|--v=*)
    logv="${i#*=}"
    shift # past argument=value
    ;;
    --alsologtostderr*)
    alsologtostderr="${i#*=}"
    shift # past argument=value
    ;;
    --modeldir=*)
    modeldir="${i#*=}"
    shift # past argument=value
    ;;
    --imgdir=*)
    imgdir="${i#*=}"
    shift # past argument=value
    ;;
    -h|--help)
    $serverbin --help
    exit 0
    ;;
    *)
            # unknown option
    ;;
esac
done

opts="--alsologtostderr=$alsologtostderr"
opts="$opts --v=$logv"
opts="$opts --modeldir=$modeldir"
opts="$opts --imgdir=$imgdir"

cmd="$serverbin $opts"
echo "$cmd"

$cmd
