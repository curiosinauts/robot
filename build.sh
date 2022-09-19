#!/bin/bash -e

set -x

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

cd $SCRIPT_DIR

gox -osarch="linux/amd64"

base_version=0.1.1

tag=$(curl -s https://api.github.com/repos/curiosinauts/platformctl/releases/latest | jq -r ".name")
tag=${tag:1}
platformctl_version="${tag}"

version=$(git log -1 --pretty=%h)

service=robot

docker build --no-cache --build-arg BASE_VERSION=${base_version} --build-arg PLATFORMCTL_VERSION=${platformctl_version} -t docker-registry.curiosityworks.org/curiosinauts/${service}:"${version}" .

docker push docker-registry.curiosityworks.org/curiosinauts/${service}:"${version}"

cat robot.tpl | sed 's/__tag__/'"${version}"'/g' > robot.yml

kubectl delete -f ./robot.yml || true

kubectl apply -f ./robot.yml

rm -f robot.yml || true

rm -f robot_linux_amd64 || true