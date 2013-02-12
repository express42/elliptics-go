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

/*
struct es3_write_result {
    struct dnet_addr *addr;
    struct dnet_cmd *cmd;
    struct dnet_addr_attr *a;
};

int dnet_write_data_wait_es3(struct dnet_session *s, struct dnet_io_control *ctl, struct es3_write_result **result);
*/
