.PHONY: test clean

MODULE := pam_oidc

$(MODULE).so: .
	go build -buildmode=c-shared -o $@

test:
	go test -v ./...

clean:
	rm -f $(MODULE).so $(MODULE).h
