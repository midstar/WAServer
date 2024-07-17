mkdir -p .test/cover
go test -coverprofile=.test/cover/c.out
go tool cover -html=.test/cover/c.out -o .test/cover/coverage.html
echo Coverage report stored in .test/cover/coverage.html