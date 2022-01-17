#! /bin/bash

set -xe

DIR=$(pwd)
TMP=$(mktemp -d)
cp ./test/testdata/griffel-config.yaml $TMP/
cp ./test/testdata/k8s-resources-pod.json $TMP/
cp ./test/testdata/k8s-resources-pod-expected.json $TMP/
cp ./test/testdata/vpa.json $TMP/
cp ./test/testdata/vpa-expected.json $TMP/
cd $TMP/

$DIR/bin/griffel --config-path griffel-config.yaml

diff k8s-resources-pod-expected.json k8s-resources-pod-generated.json
diff vpa-expected.json vpa-generated.json
