#!/bin/bash


options="--v=3 --logtostderr"
options="$options --imgdir=/tmp/imgs/"

bin=_output/inceptions
set -x
$bin $options
