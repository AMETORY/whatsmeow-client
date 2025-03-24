pkgs      = $(shell go list ./... | grep -v /tests | grep -v /vendor/ | grep -v /common/)
datetime	= $(shell date +%s)
datetimeFormat	= $(shell date +"%Y-%m-%d %H:%M:%S")

build:
	@echo "Building Go Lambda function"
	@gox -os="linux" -arch="amd64" -output="erp-wa"

deploy-staging:
	@echo "Deploying to staging"
	rsync -a erp-wa ametory@103.172.205.9:/home/ametory/erp-wa/erp-wa -v --stats --progress