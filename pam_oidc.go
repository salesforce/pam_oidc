// Copyright (c) 2021, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE.txt file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package main

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -lpam -fPIC

#include <stdlib.h>
#include <security/pam_appl.h>
#include <security/pam_modules.h>

#ifdef __linux__
#include <security/pam_ext.h>
#endif

char* argv_i(const char **argv, int i);
void pam_syslog_str(pam_handle_t *pamh, int priority, const char *str);
*/
import "C"

import (
	"context"
	"fmt"
	"log/syslog"
	"unsafe"
)

func main() {
}

//export pam_sm_authenticate_go
func pam_sm_authenticate_go(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	ctx := context.Background()

	// Copy args to Go strings
	args := make([]string, int(argc))
	for i := 0; i < int(argc); i++ {
		args[i] = C.GoString(C.argv_i(argv, C.int(i)))
	}

	// Parse config
	cfg, err := configFromArgs(args)
	if err != nil {
		pamSyslog(pamh, syslog.LOG_ERR, "failed to parse config: %v", err)
		return C.PAM_SERVICE_ERR
	}

	// Validate config
	if cfg.Issuer == "" {
		pamSyslog(pamh, syslog.LOG_ERR, "missing required option: issuer")
		return C.PAM_SERVICE_ERR
	} else if cfg.Aud == "" {
		pamSyslog(pamh, syslog.LOG_ERR, "missing required option: aud")
		return C.PAM_SERVICE_ERR
	}

	// Get (or prompt for) user
	var cUser *C.char
	if errnum := C.pam_get_user(pamh, &cUser, nil); errnum != C.PAM_SUCCESS {
		pamSyslog(pamh, syslog.LOG_ERR, "failed to get user: %v", pamStrError(pamh, errnum))
		return errnum
	}

	user := C.GoString(cUser)
	if len(user) == 0 {
		pamSyslog(pamh, syslog.LOG_WARNING, "empty user")
		return C.PAM_USER_UNKNOWN
	}

	// Get (or prompt for) password (token)
	var cToken *C.char
	if errnum := C.pam_get_authtok(pamh, C.PAM_AUTHTOK, &cToken, nil); errnum != C.PAM_SUCCESS {
		pamSyslog(pamh, syslog.LOG_ERR, "failed to get token: %v", pamStrError(pamh, errnum))
		return errnum
	}
	token := C.GoString(cToken)

	auth, err := discoverAuthenticator(ctx, cfg.Issuer, cfg.Aud)
	if err != nil {
		pamSyslog(pamh, syslog.LOG_ERR, "failed to discover authenticator: %v", err)
		return C.PAM_AUTH_ERR
	}
	auth.UserTemplate = cfg.UserTemplate
	auth.GroupsClaimKey = cfg.GroupsClaimKey
	auth.AuthorizedGroups = cfg.AuthorizedGroups

	if err := auth.Authenticate(ctx, user, token); err != nil {
		pamSyslog(pamh, syslog.LOG_WARNING, "failed to authenticate: %v", err)
		return C.PAM_AUTH_ERR
	}

	return C.PAM_SUCCESS
}

//export pam_sm_setcred_go
func pam_sm_setcred_go(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	return C.PAM_IGNORE
}

func pamStrError(pamh *C.pam_handle_t, errnum C.int) string {
	return C.GoString(C.pam_strerror(pamh, errnum))
}

func pamSyslog(pamh *C.pam_handle_t, priority syslog.Priority, format string, a ...interface{}) {
	cstr := C.CString(fmt.Sprintf(format, a...))
	defer C.free(unsafe.Pointer(cstr))

	C.pam_syslog_str(pamh, C.int(priority), cstr)
}
