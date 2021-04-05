#include <security/pam_appl.h>

#ifdef __linux__
#include <security/pam_ext.h>
#endif

// pam_sm_authenticate lightly wraps pam_sm_authenticate_go because cgo cannot
// natively create a method with 'const char**' as an argument.
int pam_sm_authenticate_go(pam_handle_t *pamh, int flags, int argc, char **argv);
int pam_sm_authenticate(pam_handle_t *pamh, int flags, int argc, const char **argv) {
  // pam_sm_authenticate_go does not modify argv, only copies them to Go strings.
  return pam_sm_authenticate_go(pamh, flags, argc, (char**)argv);
}

// pam_sm_setcred lightly wraps pam_sm_setcred_go because cgo cannot
// natively create a method with 'const char**' as an argument.
int pam_sm_setcred_go(pam_handle_t *pamh, int flags, int argc, char **argv);
int pam_sm_setcred(pam_handle_t *pamh, int flags, int argc, char **argv) {
  // pam_sm_setcred_go does not modify argv, only copies them to Go strings.
  return pam_sm_setcred_go(pamh, flags, argc, (char**)argv);
}

// argv_i returns argv[i].
char* argv_i(char **argv, int i) {
  return argv[i];
}

// pam_syslog_str logs str to pam_syslog. Calling variadic functions directly
// is not supported with cgo.
void pam_syslog_str(pam_handle_t *pamh, int priority, const char *str) {
#ifdef __linux__
  pam_syslog(pamh, priority, "%s", str);
#endif
}
