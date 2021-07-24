docker run --rm --privileged \
  -v /Users/earther/fun/0_tstream/tstream:/tstream \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /Users/earther/go/src:/go/src \
  -w /tstream \
  ghcr.io/gythialy/golang-cross:latest --snapshot --rm-dist
