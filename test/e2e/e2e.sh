#! /bin/bash

set -xe

DIR=$(pwd)
TMP=$(mktemp -d)
cp ./test/testdata/k8s-resources-pod.json $TMP/
cp ./test/testdata/k8s-resources-pod-expected.json $TMP/
cp ./test/testdata/griffel-config.yaml $TMP/
cd $TMP/

ls
$DIR/bin/griffel --config-path griffel-config.yaml
ls
diff k8s-resources-pod-expected.json k8s-resources-pod-generated.json
