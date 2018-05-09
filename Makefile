default: fmt vet errcheck test

# Taken from https://github.com/codecov/example-go#caveat-multiple-files
test:
	echo "" > coverage.txt
	for d in `go list ./... | grep -v vendor`; do \
		go test -p 1 -v -timeout 90s -race -coverprofile=profile.out -covermode=atomic $$d || exit 1; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi \
	done

vet:
	go vet ./...

errcheck:
	errcheck github.com/Shopify/sarama/...

fmt:
	@if [ -n "$$(go fmt ./...)" ]; then echo 'Please run go fmt on your code.' && exit 1; fi

install_dependencies: install_errcheck glide_install

glide_install:
	go get -u github.com/Masterminds/glide
	cd $(GOPATH)/src/github.com/remerge/sarama && glide install --strip-vendor

install_errcheck:
	go get github.com/kisielk/errcheck

get:
	go get -t

watch:
	go get github.com/cespare/reflex
	reflex -r '\.go$$' -s -- sh -c 'clear && go test -v -run=Test$(T)'

switchRemergeDep:
	go get github.com/novalagung/gorep
	gorep -path="$$GOPATH/src/github.com/remerge/sarama" -from="github.com/Shopify/sarama" -to="github.com/remerge/sarama"

switchShopifyDep:
	go get github.com/novalagung/gorep
	gorep -path="$$GOPATH/src/github.com/remerge/sarama" -from="github.com/remerge/sarama" -to="github.com/Shopify/sarama"