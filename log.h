// +build linux,cgo

#include "elliptics.h"


struct log_data_t {
	int level;
	char *msg;
	struct log_data_t *next;
};

void log_data_free(struct log_data_t *data);
struct dnet_log log_create(int level);
int log_queue_length(void);
void log_enqueue(void *priv, int level, const char *msg);
struct log_data_t *log_dequeue(void);
