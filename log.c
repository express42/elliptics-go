// +build linux,cgo

#include "log.h"

#include <pthread.h>


static pthread_mutex_t log_mutex = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t  log_available = PTHREAD_COND_INITIALIZER;
static struct log_data_t *log_queue_head;
static int log_queue_len;

void log_data_free(struct log_data_t *data) {
	free(data->msg);
	free(data);
}

struct dnet_log log_create(int level) {
	struct dnet_log res = {level, 0, log_enqueue};
	return res;
}

int log_queue_length(void) {
	int err = pthread_mutex_lock(&log_mutex);
	assert_perror(err);
	int res = log_queue_len;
	err = pthread_mutex_unlock(&log_mutex);
	assert_perror(err);
	return res;
}

void log_enqueue(void *priv, int level, const char *msg) {
	int err = pthread_mutex_lock(&log_mutex);
	assert_perror(err);

	struct log_data_t *d = malloc(sizeof(struct log_data_t));
	d->level = level;
	d->msg = strdup(msg);
	d->next = 0;
	log_queue_len++;

	// FIXME use proper queue instead
	if (log_queue_head == 0) {
		log_queue_head = d;
	} else {
		struct log_data_t *tail = log_queue_head;
		while (tail->next != 0) {
			tail = tail->next;
		}
		tail->next = d;
	}

	err = pthread_mutex_unlock(&log_mutex);
	assert_perror(err);

	err = pthread_cond_signal(&log_available);
	assert_perror(err);
}

struct log_data_t *log_dequeue(void) {
	int err = pthread_mutex_lock(&log_mutex);
	assert_perror(err);

	while (log_queue_head == 0) {
		err = pthread_cond_wait(&log_available, &log_mutex);
		assert_perror(err);
	}

	struct log_data_t *res = log_queue_head;
	log_queue_head = res->next;
	log_queue_len--;
	err = pthread_mutex_unlock(&log_mutex);
	assert_perror(err);

	return res;
}
