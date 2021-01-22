# linux build
export GOFLAGS="-mod=vendor"
export VERSION=${CI_COMMIT_TAG:=v0.0.0-build}
export GOOS="linux"
export GOARCH="amd64"
go build -ldflags "cmd.version=$VERSION" -o yala-$VERSION-$GOOS-$GOARCH

# darwin build
export GOFLAGS="-mod=vendor"
export VERSION=${CI_COMMIT_TAG:=v0.0.0-build}
export GOOS="darwin"
export GOARCH="amd64"
go build -ldflags "cmd.version=$VERSION" -o yala-$VERSION-$GOOS-$GOARCH