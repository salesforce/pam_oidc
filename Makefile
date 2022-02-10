.PHONY: test clean rpm

MODULE := pam_oidc

$(MODULE).so: .
	go build -buildmode=c-shared -o $@

rpm: $(MODULE).so
	env VERSIONED_OIDC_LIB="pam_oidc.so.$(shell hack/package_version.sh)" envsubst '$${VERSIONED_OIDC_LIB}' < src_nfpm.yaml > nfpm.yaml
	env VERSION=$(shell hack/package_version.sh) nfpm package --packager rpm

test:
	go test -v ./...

clean:
	rm -f $(MODULE).so $(MODULE).h
