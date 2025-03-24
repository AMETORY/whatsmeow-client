pkgs      = $(shell go list ./... | grep -v /tests | grep -v /vendor/ | grep -v /common/)
datetime	= $(shell date +%s)
datetimeFormat	= $(shell date +"%Y-%m-%d %H:%M:%S")

build:
	@echo "Building Go Lambda function"
	@gox -os="linux" -arch="amd64" -output="erp-wa"  