.PHONY: test clean rpm

MODULE := pam_oidc

$(MODULE).so: .
	go build -buildmode=c-shared -o $@

rpm: $(MODULE).so
	env VERSION=$(shell hack/package_version.sh) nfpm package --packager rpm

test:
	go test -v ./...

clean:
	rm -f $(MODULE).so $(MODULE).h
