############################# Main targets #############################
# Run all checks, build, and test.
install: clean staticcheck errcheck bins test
########################################################################

##### Variables ######
MAIN_FILES := $(shell find . -name "main.go")
TEST_TIMEOUT := 20s
COLOR := "\e[1;36m%s\e[0m\n"

dir_no_slash = $(patsubst %/,%,$(dir $(1)))
dirname = $(notdir $(call dir_no_slash,$(1)))
parentdirname = $(notdir $(call dir_no_slash,$(call dir_no_slash,$(1))))
define NEWLINE


endef

##### Targets ######
tidy:
	@go mod tidy

env-setup:
	@go env -w GO111MODULE=on
	@go env -w GOPROXY="https://proxy.golang.org,direct"

update-libs: env-setup
	@go get -u github.com/aws/aws-sdk-go-v2
	@go get -u github.com/aws/aws-sdk-go-v2/config
	@go get -u github.com/aws/aws-sdk-go-v2/service/sqs
	@go get -u github.com/openfaas/connector-sdk
	@go get -u github.com/sirupsen/logrus
	@make tidy

##### Targets ######
bins: env-setup
	@printf $(COLOR) "Build binaries..."
	$(foreach MAIN_FILE,$(MAIN_FILES), \
		@go build \
			-o bin/$(call parentdirname,$(MAIN_FILE))/$(call dirname,$(MAIN_FILE)) \
			$(MAIN_FILE) \
	$(NEWLINE))

test: env-setup
	@printf $(COLOR) "Run unit tests..."
	@rm -f coverage.html
	@rm -f coverage.log
	@rm -f test.log
	@go test \
		-timeout $(TEST_TIMEOUT) \
		-race \
		-coverprofile=coverage.out \
		./... | \
	tee -a test.log
	@go tool cover -html=coverage.out -o coverage.html
	@! egrep -q "^--- FAIL" test.log
	@! grep -q "no tests to run" test.log

fmtcheck: tidy
	@printf $(COLOR) "Run format check..."
	@gofmt -l $(OWN_FILES) | xargs test -z

staticcheck: env-setup
	@printf $(COLOR) "Run static check..."
	@GO111MODULE=off go get -u honnef.co/go/tools/cmd/staticcheck
	@staticcheck ./...

errcheck: env-setup
	@printf $(COLOR) "Run error check..."
	@GO111MODULE=off go get -u github.com/kisielk/errcheck
	@errcheck ./...

clean:
	rm -rf bin
