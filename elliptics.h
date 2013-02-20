// +build linux,cgo

#ifdef NDEBUG
#   error "This code uses assert() and assert_perror(), it's better not to compile with NDEBUG."
#endif

// FIXME this is bad
#define _GNU_SOURCE 1
#include <assert.h>

#include <stdlib.h>
#include <elliptics/interface.h>

#if DNET_ID_SIZE != 64
#   error "This will not work."
#endif

#if DNET_CSUM_SIZE != 64
#   error "This will not work."
#endif
